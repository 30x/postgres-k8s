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

	"github.com/spf13/cobra"
)

// masterCmd represents the master command
var masterCmd = &cobra.Command{
	Use:   "master [configure|reload]",
	Short: "Configure the master node",
	Long:  `Configure the master node`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
		}

		command := args[0]

		var err error

		if command == "configure" {
			err = ConfigureMaster(cmd, args)
		} else if command == "reload" {
			err = Reload(cmd, args)
		} else {
			err = errors.New("Unkown command")
		}

		if err != nil {
			fmt.Printf("Error: %s\n\n\n", err)

			cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(masterCmd)

	//add our two flags
	masterCmd.PersistentFlags().StringVarP(&pghbaConfLocation, "pg_hba", "", "", "The path to the pg_hba.conf file")

	masterCmd.PersistentFlags().StringVarP(&postgresConfLocation, "pg_conf", "", "", "The path to the postgres.conf file")

	masterCmd.PersistentFlags().StringVarP(&postgresDataDir, "data", "", "", "The path to the postgres data directory")

	masterCmd.PersistentFlags().StringVarP(&hostname, "hostname", "", "", "The hostname of the current machine")

}

//ConfigureMaster configure this node as a master
func ConfigureMaster(cmd *cobra.Command, args []string) error {

	inputErrors := &InputErrors{}

	//sanity check
	if pghbaConfLocation == "" {
		inputErrors.Append(errors.New("You must specify a pg_hba.conf file"))
	}

	if postgresConfLocation == "" {
		inputErrors.Append(errors.New("You must specify a postgres.conf file"))
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

	pgaHbaFile, err := os.OpenFile(pghbaConfLocation, os.O_RDWR|os.O_APPEND, 0660)

	if err != nil {
		return err
	}

	postgresConf, err := os.OpenFile(postgresConfLocation, os.O_RDWR|os.O_APPEND, 0660)

	if err != nil {
		return err
	}
	//TODO, this causes a race condition, since the second pod is not yet up and running.  This needs to be better
	// slaveHostname, err := GetPetPodNameAtIndex(hostname, 1)

	// if err != nil {
	// 	return err
	// }

	slaveHostname := "10.244.0.0/16"

	fmt.Println("Configuring hba conf file")

	//TODO, this needs to have a username and password
	hbaTemplate := `

	
host	replication	postgres	%s	trust
`

	outputLine := fmt.Sprintf(hbaTemplate, slaveHostname)

	fmt.Printf("Adding the line %s to the file %s", outputLine, pgaHbaFile.Name())

	pgaHbaFile.WriteString(outputLine)

	err = pgaHbaFile.Sync()

	if err != nil {
		return err
	}

	err = pgaHbaFile.Close()

	if err != nil {
		return err
	}

	//now update the postgres file

	fmt.Println("Creating archive directory")

	postgresArchiveDir := GetPostgresArchivepath(postgresDataDir)

	err = os.MkdirAll(postgresArchiveDir, 0700)

	if err != nil {
		return err
	}

	fmt.Println("Configuring postgres conf file")

	postgresTemplate := `
wal_level = hot_standby
archive_mode = on
archive_command = 'test ! -f %s/%%f && cp %%p %s/%%f'
max_wal_senders = 3
synchronous_standby_names = '%s'
	`

	outputLine = fmt.Sprintf(postgresTemplate, postgresArchiveDir, postgresArchiveDir, slaveHostname)

	postgresConf.WriteString(outputLine)

	err = postgresConf.Sync()

	if err != nil {
		return err
	}

	err = postgresConf.Close()

	if err != nil {
		return err
	}
	return nil
}
