package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

//GetPetPodNameAtIndex Get a pod hostname at the index specified
func GetPetPodNameAtIndex(hostname string, index int) (string, error) {
	if hostname == "" {
		return "", errors.New("You must specify a hostname")
	}

	hostNameOnly := ParseHostnameFromFQDN(hostname)

	parts := strings.Split(hostNameOnly, "-")

	if len(parts) != 2 {
		return "", errors.New("Unkown format encountered, expected 2 parts when split with the '-' char")
	}

	podName := fmt.Sprintf("%s-%d", parts[0], index)

	fqdn := ParseFQDN(hostname)

	//we have a full domain, append it
	if fqdn != "" {
		podName = podName + "." + fqdn
	}

	return podName, nil
}

//ParseHostnameFromFQDN parse out the hosthame from the FQDN
func ParseHostnameFromFQDN(hostname string) string {

	//split the hostname and get the host
	index := strings.Index(hostname, ".")

	if index == -1 {
		return hostname
	}

	return hostname[:index]

}

//ParseFQDN parse the FQDN from a full hostname
func ParseFQDN(hostname string) string {
	index := strings.Index(hostname, ".")

	if index == -1 {
		return ""
	}

	return hostname[index+1:]
}

//DirectoryExists returns true of false if the path exists
func DirectoryExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err)
	}

	return stat.IsDir()
}

//RemoveDirContents remove the directory contents
func RemoveDirContents(dir string) error {
	stat, err := os.Stat(dir)

	if err != nil {
		return err
	}

	mode := stat.Mode()

	err = os.RemoveAll(dir)

	if err != nil {
		return err
	}

	return os.MkdirAll(dir, mode)

}

//ConfigureReplica configure this node as a master
func ConfigureReplica(cmd *cobra.Command, args []string, walLocator WALLocator) error {

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

	masterHostname, err := walLocator.GetHostName()
	// masterHostname, err := GetPetPodNameAtIndex(hostname, 0)

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

	//now that the pre-backup has executed, the container will start the postgres process

	return nil
}
