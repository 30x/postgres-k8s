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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
)

var deletePvc bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an existing cluster.",
	Long: `This will delete an existing cluster.  If this command is run without the -d parameters, the cluster can be re-started via the create command using the same name.  
If -d is specified all persistent disks will be removed.  Use this with care, you cannot recover from this operation`,
	Run: func(cmd *cobra.Command, args []string) {
		errors := &InputErrors{}

		if clusterName == "" {
			errors.Add("ERROR: clusterName is a required parameter")
		}

		if namespace == "" {
			errors.Add("ERROR: namespace is a required parameter")
		}

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := DeleteCluster(clusterName, namespace, deletePvc)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&clusterName, "clusterName", "c", "", "The cluster name to create.")

	deleteCmd.Flags().BoolVarP(&deletePvc, "deletePvc", "d", false, "Delete the Persistent Volume Claims.  Use with care, this will destroy your data")

}

//DeleteCluster delete the cluster
func DeleteCluster(clusterName, namespace string, deletePVC bool) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	err = deleteServices(client, clusterName, namespace)

	if err != nil {
		return err
	}

	err = deleteReplicaSets(client, clusterName, namespace)

	if err != nil {
		return err
	}

	err = deleteReplicaPods(client, clusterName, namespace)

	if err != nil {
		return err
	}

	if deletePVC {
		err = deletePersistentVolumeClaims(client, clusterName, namespace)

		if err != nil {
			return err
		}
	}

	return nil

}

func deletePersistentVolumeClaims(client *kubernetes.Clientset, clusterName, namespace string) error {

	selector := createClusterSelector(clusterName)

	err := client.PersistentVolumeClaims(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{
		LabelSelector: selector,
	})

	return err

}

func deleteReplicaSets(client *kubernetes.Clientset, clusterName, namespace string) error {

	selector := createClusterSelector(clusterName)

	err := client.ReplicaSets(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{
		LabelSelector: selector,
	})

	return err
}

func deleteReplicaPods(client *kubernetes.Clientset, clusterName, namespace string) error {

	selector := createClusterSelector(clusterName)

	err := client.Pods(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{
		LabelSelector: selector,
	})

	return err
}

func deleteServices(client *kubernetes.Clientset, clusterName, namespace string) error {

	selector := createClusterSelector(clusterName)

	err := client.Services(namespace).DeleteCollection(&v1.DeleteOptions{}, v1.ListOptions{
		LabelSelector: selector,
	})

	return err
}
