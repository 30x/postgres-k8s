package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

//WALLocator an interface to receive WAL data
type WALLocator struct {
	//GetHostName get the host name to receive WAL data from
	GetHostName func() (string, error)
}

//ConfigureFromBackup configure this node as a master
func ConfigureFromBackup(cmd *cobra.Command, args []string, walLocator *WALLocator) error {

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

	if postgresConfLocation == "" {
		inputErrors.Append(errors.New("You must specify a postgres.conf file"))
	}

	if inputErrors.HasErrors() {
		return inputErrors
	}

	backupDir, err := exec.LookPath("pg_basebackup")

	if err != nil {
		return err
	}

	masterHostname, err := walLocator.GetHostName()

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

	postgresConf, err := os.OpenFile(postgresConfLocation, os.O_RDWR|os.O_APPEND, 0660)

	if err != nil {
		return err
	}

	postgresTemplate := `
hot_standby = on
	`

	postgresConf.WriteString(postgresTemplate)

	err = postgresConf.Sync()

	if err != nil {
		return err
	}

	err = postgresConf.Close()

	if err != nil {
		return err
	}

	//make sure postgres owns the data directory

	//now that the pre-backup has executed, the container will start the postgres process
	return nil

}

//Reload reload postgres
func Reload(cmd *cobra.Command, args []string) error {

	// fmt.Println("master called")

	// fmt.Printf("command is %+v\n", cmd)

	// fmt.Printf("ares are %+v\n", args)

	return nil
}

//GetPostgresArchivepath get the path to the postgres archive
func GetPostgresArchivepath(postgresDataDir string) string {
	return filepath.Join(postgresDataDir, "archive")
}
