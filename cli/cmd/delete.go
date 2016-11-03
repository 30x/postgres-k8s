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

	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/labels"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
)

var deleteArgs *DeleteArgs

//DeleteArgs the args for the create command
type DeleteArgs struct {
	namespace   string
	clusterName string
	deletePvc   bool
}

func (args *DeleteArgs) validate() *InputErrors {

	errors := &InputErrors{}

	if args.clusterName == "" {
		errors.Add("ERROR: clusterName is a required parameter")
	}

	if args.namespace == "" {
		errors.Add("ERROR: namespace is a required parameter")
	}

	return errors

}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing cluster.",
	Long: `This will delete an existing cluster.  If this command is run without the -d parameters, the cluster can be re-started via the create command using the same name.  
If -d is specified all persistent disks will be removed.  Use this with care, you cannot recover from this operation`,
	Run: func(cmd *cobra.Command, args []string) {
		errors := deleteArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := DeleteCluster(deleteArgs.namespace, deleteArgs.clusterName, deleteArgs.deletePvc)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)

	deleteArgs = &DeleteArgs{}

	deleteCmd.Flags().StringVarP(&deleteArgs.namespace, "namespace", "n", "", "The namespace to use")
	deleteCmd.Flags().StringVarP(&deleteArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")

	deleteCmd.Flags().BoolVarP(&deleteArgs.deletePvc, "deletePvc", "d", false, "Delete the Persistent Volume Claims.  Use with care, this will destroy your data")

}

//DeleteCluster delete the cluster
func DeleteCluster(namespace, clusterName string, deletePVC bool) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	err = deleteServices(client, namespace, clusterName)

	if err != nil {
		return err
	}

	err = deleteReplicaSets(client, namespace, clusterName)

	if err != nil {
		return err
	}

	err = deleteReplicaPods(client, namespace, clusterName)

	if err != nil {
		return err
	}

	if deletePVC {
		err = deletePersistentVolumeClaims(client, namespace, clusterName)

		if err != nil {
			return err
		}
	}

	return nil

}

func deletePersistentVolumeClaims(client *kubernetes.Clientset, namespace, clusterName string) error {

	fmt.Println("Deleting persistent volume claims")

	selector, err := labels.Parse(createClusterSelector(clusterName))

	if err != nil {
		return err
	}

	err = client.PersistentVolumeClaims(namespace).DeleteCollection(&api.DeleteOptions{}, api.ListOptions{
		LabelSelector: selector,
	})

	return err

}

func deleteReplicaSets(client *kubernetes.Clientset, namespace, clusterName string) error {

	fmt.Println("Deleting Replica Sets")

	selector, err := labels.Parse(createClusterSelector(clusterName))

	if err != nil {
		return err
	}

	err = client.ReplicaSets(namespace).DeleteCollection(&api.DeleteOptions{}, api.ListOptions{
		LabelSelector: selector,
	})

	return err
}

func deleteReplicaPods(client *kubernetes.Clientset, namespace, clusterName string) error {

	fmt.Println("Deleting Replica Pods")

	selector, err := labels.Parse(createClusterSelector(clusterName))

	if err != nil {
		return err
	}

	err = client.Pods(namespace).DeleteCollection(&api.DeleteOptions{}, api.ListOptions{
		LabelSelector: selector,
	})

	return err
}

func deleteServices(client *kubernetes.Clientset, namespace, clusterName string) error {

	fmt.Println("Deleting Services")

	selector, err := labels.Parse(createClusterSelector(clusterName))

	if err != nil {
		return err
	}

	services, err := client.Services(namespace).List(api.ListOptions{
		LabelSelector: selector,
	})

	for _, service := range services.Items {
		err := client.Services(namespace).Delete(service.Name, nil)

		if err != nil {
			return err
		}
	}

	return err
}
