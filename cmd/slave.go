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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// slaveCmd represents the master command
var slaveCmd = &cobra.Command{
	Use:   "slave [configure|reload]",
	Short: "Configure the slave node",
	Long:  `Configure the slave node`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		command := args[0]

		var err error

		if command == "configure" {

			walLocator := &WALLocator{
				GetHostName: func() (string, error) {
					return GetPetPodNameAtIndex(hostname, 0)
				},
			}

			err = ConfigureFromBackup(cmd, args, walLocator)
		} else if command == "reload" {
			err = Reload(cmd, args)
		} else {
			err = errors.New("Unkown command")
		}

		if err != nil {
			// cmd.std
			// cmd.OutOrStderr().Write([]byte(err.Error()))

			fmt.Printf("Error: %s\n\n\n", err)

			cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(slaveCmd)

	//add our two flags
	slaveCmd.PersistentFlags().StringVarP(&postgresDataDir, "data", "", "", "The path to the postgres data directory")

	slaveCmd.PersistentFlags().StringVarP(&hostname, "hostname", "", "", "The hostname of the current machine")

	slaveCmd.PersistentFlags().StringVarP(&postgresPort, "port", "", "", "The postgres port to run a backup")

	slaveCmd.PersistentFlags().StringVarP(&postgresUser, "user", "", "", "The postgres user to restore the backup")

	slaveCmd.PersistentFlags().StringVarP(&postgresConfLocation, "pg_conf", "", "", "The path to the postgres.conf file")

}
