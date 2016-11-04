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

var setReplicasArgs *SetReplicasArgs

//SetReplicasArgs the args for the create command
type SetReplicasArgs struct {
	namespace   string
	clusterName string
}

func (args *SetReplicasArgs) validate() *InputErrors {

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
var setReplicaCmd = &cobra.Command{
	Use:   "setreplicas",
	Short: "Sets the replicas into the pod and restarts the pg process ",
	Long:  `This will set the replicas manually if they have been disabled for any reason`,
	Run: func(cmd *cobra.Command, args []string) {

		errors := setReplicasArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := SetReplicaNames(setReplicasArgs.namespace, setReplicasArgs.clusterName)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(setReplicaCmd)

	setReplicasArgs = &SetReplicasArgs{}

	setReplicaCmd.Flags().StringVarP(&setReplicasArgs.namespace, "namespace", "n", "", "The namespace to use")
	setReplicaCmd.Flags().StringVarP(&setReplicasArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")

}

//SetReplicaNames Set the replica names back into the master
func SetReplicaNames(namespace, clusterName string) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	return updateMasterWithReplicas(client, namespace, clusterName)

}
