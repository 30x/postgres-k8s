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
	"log"
	"os"
	"time"

	api "k8s.io/client-go/1.4/pkg/api"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
)

var failoverArgs *FailoverArgs

//FailoverArgs the args for the create command
type FailoverArgs struct {
	namespace   string
	clusterName string
	force       bool
}

func (args *FailoverArgs) validate() *InputErrors {

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
var failoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Sets the first running replica to the master",
	Long:  `This will set the first running replica to the master.  Note that if the master is already detected as running, by default this will not allow you to override the master without the force option`,
	Run: func(cmd *cobra.Command, args []string) {
		errors := failoverArgs.validate()

		if errors.HasErrors() {
			fmt.Printf("\n")
			fmt.Fprint(os.Stderr, errors.Error())
			fmt.Printf("ERROR: Unable to execute command, see usage below\n\n")
			cmd.Help()
			return
		}

		err := Failover(failoverArgs.namespace, failoverArgs.clusterName, failoverArgs.force)

		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(failoverCmd)

	failoverArgs = &FailoverArgs{}

	failoverCmd.Flags().StringVarP(&failoverArgs.namespace, "namespace", "n", "", "The namespace to use")
	failoverCmd.Flags().StringVarP(&failoverArgs.clusterName, "clusterName", "c", "", "The cluster name to create.")
	failoverCmd.Flags().BoolVarP(&failoverArgs.force, "force", "f", false, "Force the failover, even if a master is detected")
}

//Failover delete the cluster
func Failover(namespace, clusterName string, force bool) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	oldMasterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil && !isNotFoundError(err) {
		return err
	}

	if oldMasterPod != nil && !force {
		return fmt.Errorf("Master pod '%s' is already running.  If you want to failover, you must specify the force flag.", oldMasterPod.Name)
	}

	//it's not found, continue
	replicas, err := getReplicaPods(client, namespace, clusterName)

	if err != nil {
		return err
	}

	if len(replicas) < 1 {
		return fmt.Errorf("No replica pods could be found")
	}

	newMaster := replicas[0]

	ownerRefs := newMaster.OwnerReferences

	if len(ownerRefs) != 1 {
		return fmt.Errorf("Expected the pod %s to only have 1 owner.  Short circuiting", newMaster.Name)
	}

	// kind := ownerRefs[0].Kind
	name := ownerRefs[0].Name

	rs, err := client.ReplicaSets(namespace).Get(name)

	if err != nil {
		return err
	}

	rs.Spec.Template.Labels["master"] = "true"

	updatedRS, err := client.ReplicaSets(namespace).Update(rs)

	if err != nil {
		return err
	}

	log.Printf("Updated Replica Set %s with new master labels", updatedRS.Name)

	//now update the master

	newMaster.Labels["master"] = "true"

	updatedPod, err := client.Pods(namespace).Update(&newMaster)

	if err != nil {
		return err
	}

	log.Printf("Updated pod %s with master labels", updatedPod.Name)

	log.Printf("Re-writing synchrnous replicas")

	containerName, err := getContainerNameForPod(updatedPod)

	if err != nil {
		return err
	}

	indexes, err := getActiveIndexesForCluster(replicas[1:])

	if err != nil {
		return err
	}

	err = updateMasterConfigurationWithReplicaIds(namespace, updatedPod.Name, containerName, indexes)

	if err != nil {
		return err
	}

	log.Printf("Bouncing pod for service discovery")

	//now kill the pod and get the master
	err = client.Pods(namespace).Delete(updatedPod.Name, &api.DeleteOptions{})

	if err != nil {
		return err
	}

	log.Printf("Waiting for pod to start")

	//wait for the pods to come back up
	err = waitForPodsToStart(client, namespace, clusterName, len(replicas), 60*time.Second)

	if err != nil {
		return err
	}

	masterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil {
		return err
	}

	log.Printf("Signaling to replica it needs to become the master")

	command := []string{"/usr/bin/touch", "/tmp/postgresql.trigger.5432"}

	_, stderr, err := k8s.ExecCommandWithStdoutStderr(namespace, masterPod.Name, containerName, command)

	if err != nil {
		return err
	}

	if len(stderr) != 0 {
		return fmt.Errorf("An error occured when running the touch command on postgres.  The stderr output is below.  \n %s", stderr)
	}

	//TODO occasionally this hangs.  Not sure why, seems to be service switch issue with k8s. Killing the pod and respawing seems to resolve.  We might want to do that before signaling

	//test the new master
	err = testMaster(client, namespace, clusterName)

	if err != nil {
		return err
	}

	log.Printf("Postgres is now online and ready to rock!")

	return nil

}
