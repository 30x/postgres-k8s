package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
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

	return podName, nil
}

//ParseHostnameFromFQDN parse out the hosthame from the FQDN
func ParseHostnameFromFQDN(hostname string) string {

	//split the hostname and get the host
	parts := strings.Split(hostname, ".")

	return parts[0]

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
