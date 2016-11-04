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

var getMasterArgs *GetMasterArgs

//GetMasterArgs the args for the create command
type GetMasterArgs struct {
	namespace   string
	clusterName string
}

func (args *GetMasterArgs) validate() *InputErrors {

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
var getMasterCmd = &cobra.Command{
	Use:   "get master",
	Short: "Get the master of this cluster",
	Long:  `This will return the currently running pod that is the master node`,
	Run: func(cmd *cobra.Command, args []string) {
		errors := getMasterArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := PrintMaster(getMasterArgs.namespace, getMasterArgs.clusterName)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(getMasterCmd)

	getMasterArgs = &GetMasterArgs{}

	getMasterCmd.Flags().StringVarP(&getMasterArgs.namespace, "namespace", "n", "", "The namespace to use")
	getMasterCmd.Flags().StringVarP(&getMasterArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")
}

//PrintMaster delete the cluster
func PrintMaster(namespace, clusterName string) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	masterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil && !isNotFoundError(err) {
		return err
	}

	if masterPod == nil {
		fmt.Printf("Could not find the master pod, it does not appear to be running.")
	}

	fmt.Printf("Master pod is '%s'", masterPod.Name)

	return nil
}
