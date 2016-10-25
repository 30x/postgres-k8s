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

var clusterName string
var storageClass string
var numReplicas int

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster or restart a cluster with existing persistent volume claims.",
	Long: `This will spin up a new cluster.  In the event no persistent volume claims, services, or replica sets exist, they will be created.
	This command is intentionally idempotent, so it can be run multiple times, and any missing resource will be created.  If the Persistent Volume Claims
	exist, such as after running delete without the -d parameters, existing cluster data and configuration will be used`,

	Run: func(cmd *cobra.Command, args []string) {

		errors := &InputErrors{}

		if clusterName == "" {
			errors.Add("clusterName is a required parameter")
		}

		if namespace == "" {
			errors.Add("namespace is a required parameter")
		}

		if errors.HasErrors() {
			fmt.Fprint(os.Stderr, errors.Error())
			return
		}

		err := createCluster(clusterName, storageClass, numReplicas)

		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&clusterName, "clusterName", "c", "", "The cluster name to create.")
	createCmd.Flags().StringVarP(&storageClass, "storageClass", "s", "postgresv1", "The storage class to use when creating the cluster. Defaults to 'postgresv1'")
	createCmd.Flags().IntVarP(&numReplicas, "numReplicas", "n", 2, "The number of replicas to create.  Defaults to 2.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

//createCluster Create a cluster and all of it's resources
func createCluster(clusterName, storageClassName string, numReplicas int) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	storageClass, err := client.StorageClasses().Get(storageClassName)

	if err != nil {
		return err
	}

	if storageClass == nil {
		storageClass = k8s.CreateStorageClass(storageClassName)

		storageClass, err = client.StorageClasses().Create(storageClass)

		if err != nil {
			return err
		}
	}

	return nil

}
