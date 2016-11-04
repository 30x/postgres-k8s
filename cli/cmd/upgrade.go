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
	"time"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
	api "k8s.io/client-go/1.4/pkg/api"
)

var upgradeArgs *UpgradeArgs

//UpgradeArgs the args for the create command
type UpgradeArgs struct {
	namespace   string
	clusterName string
	imageName   string
}

func (args *UpgradeArgs) validate() *InputErrors {

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
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Updates existing pods for a new docker image ",
	Long:  `This will upgrade pods to a new docker image by updating the Replica Set, then killing the pod.  Note that this is only indended to perform minor upgrade`,
	Run: func(cmd *cobra.Command, args []string) {

		errors := upgradeArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := UpgradeReplicaSets(upgradeArgs.namespace, upgradeArgs.clusterName, upgradeArgs.imageName)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(upgradeCmd)

	upgradeArgs = &UpgradeArgs{}

	upgradeCmd.Flags().StringVarP(&upgradeArgs.namespace, "namespace", "n", "", "The namespace to use")
	upgradeCmd.Flags().StringVarP(&upgradeArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")

	upgradeCmd.Flags().StringVarP(&upgradeArgs.imageName, "image", "i", k8s.DefaultImage, fmt.Sprintf("The image name to use when creating the cluster. Defaults to '%s'", k8s.DefaultImage))

}

//UpgradeReplicaSets add the number of replicas to the cluster in the given namespace
func UpgradeReplicaSets(namespace, clusterName, imageName string) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	replicas, err := getClusterReplicaSets(client, namespace, clusterName)

	if err != nil {
		return err
	}

	size := len(replicas)

	for _, replica := range replicas {

		replica.Spec.Template.Spec.Containers[0].Image = imageName

		log.Printf("Upgrading replica set '%s'", replica.Name)

		updatedReplica, err := client.ReplicaSets(namespace).Update(&replica)

		if err != nil {
			return err
		}

		//now get the currently running pod
		pods, err := getPodsForReplicaSet(client, updatedReplica)

		if err != nil {
			return err
		}

		for _, pod := range pods {
			log.Printf("Delete")
			err = client.Pods(namespace).Delete(pod.Name, &api.DeleteOptions{})

			if err != nil {
				return err
			}
		}

		//wait for the pods to come back up before rolling to the next one
		log.Printf("Waiting up to 5 minutes for pod to restart")

		waitForPodsToStart(client, namespace, clusterName, size, 5*time.Minute)

		err = testMaster(client, namespace, clusterName)

		if err != nil {
			return err
		}
	}

	return nil

}
