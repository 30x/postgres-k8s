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
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var postgresPort string
var postgresUser string

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
			err = ConfigureSlave(cmd, args)
		} else if command == "reload" {
			err = Reload(cmd, args)
		} else {
			err = errors.New("Unkown command")
		}

		if err != nil {
			// cmd.std
			// cmd.OutOrStderr().Write([]byte(err.Error()))

			fmt.Printf("Error: %s", err)

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

}

//ConfigureSlave configure this node as a master
func ConfigureSlave(cmd *cobra.Command, args []string) error {

	inputErrors := &InputErrors{}

	//sanity check
	if postgresPort == "" {
		inputErrors.Append(errors.New("You must specify a postgres port "))
	}

	if postgresUser == "" {
		inputErrors.Append(errors.New("You must specify the postgres user"))
	}

	if postgresDataDir == "" {
		inputErrors.Append(errors.New("You must specify a postgres data directory"))
	}

	if hostname == "" {
		inputErrors.Append(errors.New("You must specify a hostname"))
	}

	if inputErrors.HasErrors() {
		return inputErrors
	}

	backupDir, err := exec.LookPath("pg_basebackup")

	if err != nil {
		return err
	}

	masterHostname, err := GetPetPodNameAtIndex(hostname, 0)

	if err != nil {
		return err
	}

	fmt.Printf("Restoring a backup from host %s\n", masterHostname)

	//check the data dir
	if DirectoryExists(postgresDataDir) {
		//it exists, remove it's contents and start streaming
		RemoveDirContents(postgresDataDir)
	}

	//run the command (as postgres)
	backupCommand := exec.Command(backupDir, "-D", postgresDataDir, "-p", postgresPort, "-U", postgresUser, "-v", "-h", masterHostname, "--xlog-method=stream")

	backupCommand.Stderr = os.Stderr
	backupCommand.Stdout = os.Stdout

	err = backupCommand.Run()

	if err != nil {
		return err
	}

	return nil
}
