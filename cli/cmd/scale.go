// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
)

var scaleArgs *ScaleArgs

//ScaleArgs the args for the create command
type ScaleArgs struct {
	namespace      string
	clusterName    string
	numReplicas    int
	storageClass   string
	diskSizeInGigs int
}

func (args *ScaleArgs) validate() *InputErrors {

	errors := &InputErrors{}

	if args.clusterName == "" {
		errors.Add("ERROR: clusterName is a required parameter")
	}

	if args.namespace == "" {
		errors.Add("ERROR: namespace is a required parameter")
	}

	return errors

}

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Adds replicas to the cluster",
	Long: `This will add replicas to an existing cluster.  A typical use might be to increase read capacity, or to add capacity after a failover.  
Another potential use is to add capacity with a new transicator instance, such as a major Postgres version upgrade.  Once these new nodes are current, the failover command can be run to the new nodes, and the old nodes can be removed`,
	Run: func(cmd *cobra.Command, args []string) {

		errors := scaleArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := AddReplicas(scaleArgs.namespace, scaleArgs.clusterName, scaleArgs.storageClass, scaleArgs.numReplicas, scaleArgs.diskSizeInGigs)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(scaleCmd)

	scaleArgs = &ScaleArgs{}

	scaleCmd.Flags().StringVarP(&scaleArgs.namespace, "namespace", "n", "", "The namespace to use")
	scaleCmd.Flags().StringVarP(&scaleArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")
	scaleCmd.Flags().IntVarP(&scaleArgs.numReplicas, "replicas", "r", 1, "The number of replicas to create.  Defaults to 1.")
	scaleCmd.Flags().StringVarP(&scaleArgs.storageClass, "storageClass", "s", "postgresv1", "The storage class to use when creating the cluster. Defaults to 'postgresv1'")
	scaleCmd.Flags().IntVarP(&scaleArgs.diskSizeInGigs, "diskSize", "d", 250, "The size of the EBS volume to allocate.  Defaults to 250 GB.  Unit is an int64")

}

//AddReplicas add the number of replicas to the cluster in the given namespace
func AddReplicas(namespace, clusterName, storageClass string, numReplicas, sidkSizeInGigs int) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	replicas, err := getReplicaPods(client, namespace, clusterName)

	if err != nil {
		return err
	}

	masterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil {
		return err
	}

	indexes, err := getActiveIndexesForCluster(append(replicas, *masterPod))

	if err != nil {
		return err
	}

	//get the max index, and increment it one.
	startIndex := indexes[len(indexes)-1] + 1

	endIndex := startIndex + numReplicas

	for i := startIndex; i < endIndex; i++ {
		newClaim := k8s.CreatePersistentVolumeClaim(clusterName, storageClass, i, sidkSizeInGigs)

		_, err = checkAndCreatePVC(client, namespace, newClaim)

		if err != nil {
			return err
		}

		replica := k8s.CreateReplica(clusterName, i)

		_, err = checkAndCreateReplicaSet(client, namespace, replica)

		if err != nil {
			return err
		}

	}

	return updateMasterWithReplicas(client, namespace, clusterName)

}
