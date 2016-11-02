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
	"log"
	"os"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
)

var failoverArgs *FailoverArgs

//FailoverArgs the args for the create command
type FailoverArgs struct {
	namespace   string
	clusterName string
	force       bool
}

func (args *FailoverArgs) validate() *InputErrors {

	errors := &InputErrors{}

	if args.clusterName == "" {
		errors.Add("ERROR: clusterName is a required parameter")
	}

	if args.namespace == "" {
		errors.Add("ERROR: namespace is a required parameter")
	}

	return errors

}

// failoverCmd represents the failover command
var failoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		errors := failoverArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := Failover(failoverArgs.namespace, failoverArgs.clusterName, failoverArgs.force)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(failoverCmd)

	failoverArgs = &FailoverArgs{}

	failoverCmd.Flags().StringVarP(&failoverArgs.namespace, "namespace", "n", "", "The namespace to use")
	failoverCmd.Flags().StringVarP(&failoverArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")
	failoverCmd.Flags().BoolVarP(&failoverArgs.force, "force", "f", false, "Force the failover, even if a master is detected")
}

//Failover delete the cluster
func Failover(namespace, clusterName string, force bool) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	oldMasterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil && !isNotFoundError(err) {
		return err
	}

	if oldMasterPod != nil && !force {
		return fmt.Errorf("Master pod is already running.  If you want to failover, you must specify the force flag.")
	}

	//it's not found, continue
	replicas, err := getReplicaPods(client, namespace, clusterName)

	if err != nil {
		return err
	}

	if len(replicas) < 1 {
		return fmt.Errorf("No replica pods could be found")
	}

	newMaster := replicas[0]

	newMaster.Labels["role"] = "master"

	//patch the master with the labels

	// updatedMasterPod, err := client.Pods(namespace).Patch()

	// if err != nil {
	// 	return err
	// }

	//TODO update the replica set in case the pod dies

	// k8s.CreateMaster(clusterName, rep)

	ownerRefs := newMaster.OwnerReferences

	if len(ownerRefs) != 1 {
		return fmt.Errorf("Expected the pod %s to only have 1 owner.  Short circuiting", newMaster.Name)
	}

	// kind := ownerRefs[0].Kind
	name := ownerRefs[0].Name

	rs, err := client.ReplicaSets(namespace).Get(name)

	if err != nil {
		return err
	}

	rs.Spec.Template.Labels["role"] = "master"

	updatedRS, err := client.ReplicaSets(namespace).Update(rs)

	if err != nil {
		return err
	}

	log.Printf("Updated Replica Set %s with new master labels", updatedRS.Name)

	// log.Printf("Updated pod %s with master labels", updatedMasterPod.Name)

	return nil

}
