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
	"strings"

	"strconv"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/30x/postgres-k8s/cli/k8s"
	"github.com/spf13/cobra"
	extv1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

var clusterName string
var storageClass string
var numReplicas int
var diskSizeInGigs int

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

		err := CreateCluster(clusterName, storageClass, numReplicas, diskSizeInGigs)

		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&clusterName, "clusterName", "c", "", "The cluster name to create.")
	createCmd.Flags().StringVarP(&storageClass, "storageClass", "s", "postgresv1", "The storage class to use when creating the cluster. Defaults to 'postgresv1'")
	createCmd.Flags().IntVarP(&numReplicas, "replicas", "r", 2, "The number of replicas to create.  Defaults to 2.")
	createCmd.Flags().IntVarP(&diskSizeInGigs, "diskSize", "d", 250, "The size of the EBS volume to allocate.  Defaults to 250 GB.  Unit is an int64")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

//CreateCluster Create a cluster and all of it's resources
func CreateCluster(clusterName, storageClassName string, numReplicas, diskSizeInGigs int) error {
	client, err := k8s.CreateClientFromEnv()

	if err != nil {
		return err
	}

	log.Printf("Checking for storage class %s", storageClassName)

	storageClass, err := client.StorageClasses().Get(storageClassName)

	if err != nil {
		if !isNotFoundError(err) {
			return err
		}
		//no storage class, create it

		log.Printf("Storage class %s does not exist, creating", storageClassName)
		storageClass = k8s.CreateStorageClass(storageClassName)

		storageClass, err = client.StorageClasses().Create(storageClass)

		if err != nil {
			return err
		}

		log.Printf("Storage class %s created", storageClassName)
	}

	//now create the services
	writeService := k8s.CreateWriteService(clusterName)
	readService := k8s.CreateReadService(clusterName)

	writeService, err = checkAndCreateService(client, writeService)

	if err != nil {
		return err
	}

	readService, err = checkAndCreateService(client, readService)

	if err != nil {
		return err
	}

	//spin up the replicas.  We start at 1 b/c index 0 is always the master

	replicaIds := []string{}

	for i := 1; i < numReplicas+1; i++ {

		pvc := k8s.CreatePersistentVolumeClaim(clusterName, storageClassName, i, diskSizeInGigs)

		//now create the replica
		rs := k8s.CreateReplica(clusterName, i)

		pvc, err := checkAndCreatePVC(client, pvc)

		if err != nil {
			return err
		}

		rs, err = checkAndCreateReplicaSet(client, rs)

		if err != nil {
			return err
		}

		replicaIds = append(replicaIds, strconv.Itoa(i))
	}

	//now create a master

	masterPvc := k8s.CreatePersistentVolumeClaim(clusterName, storageClassName, 0, diskSizeInGigs)

	masterRs := k8s.CreateMaster(clusterName, replicaIds)

	masterPvc, err = checkAndCreatePVC(client, masterPvc)

	if err != nil {
		return err
	}

	masterRs, err = checkAndCreateReplicaSet(client, masterRs)

	if err != nil {
		return err
	}

	//now check and validate service

	return nil

}

//either returns the existing pvc, or creates a new one and returns the recreated resource
func checkAndCreatePVC(client *kubernetes.Clientset, pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	log.Printf("Checking for pvc %s", pvc.Name)
	k8sPvc, err := client.PersistentVolumeClaims(namespace).Get(pvc.Name)

	if err != nil {

		if !isNotFoundError(err) {
			return nil, err
		}

		log.Printf("PVC %s does not exist, creating", pvc.Name)
		k8sPvc, err = client.PersistentVolumeClaims(namespace).Create(pvc)

		if err != nil {
			return nil, err
		}

		log.Printf("PVC %s created", pvc.Name)

	}

	return k8sPvc, nil

}

//either returns the existing pvc, or creates a new one and returns the recreated resource
func checkAndCreateService(client *kubernetes.Clientset, pvc *v1.Service) (*v1.Service, error) {
	log.Printf("Checking for service %s", pvc.Name)
	k8sPvc, err := client.Services(namespace).Get(pvc.Name)

	if err != nil {

		if !isNotFoundError(err) {
			return nil, err
		}

		log.Printf("Service %s does not exist, creating", pvc.Name)
		k8sPvc, err = client.Services(namespace).Create(pvc)

		if err != nil {
			return nil, err
		}

		log.Printf("Service %s created", pvc.Name)

	}

	return k8sPvc, nil

}

//either returns the existing pvc, or creates a new one and returns the recreated resource
func checkAndCreateReplicaSet(client *kubernetes.Clientset, rs *extv1beta1.ReplicaSet) (*extv1beta1.ReplicaSet, error) {
	log.Printf("Checking for ReplicaSet %s", rs.Name)
	k8sRS, err := client.ReplicaSets(namespace).Get(rs.Name)

	if err != nil {
		if !isNotFoundError(err) {
			return nil, err
		}

		log.Printf("ReplicaSet %s does not exist, creating", rs.Name)
		k8sRS, err = client.ReplicaSets(namespace).Create(rs)

		if err != nil {
			return nil, err
		}
		log.Printf("ReplicaSet %s created", rs.Name)

	}

	return k8sRS, nil

}

//isNotFoundError Returns true if the resource is not found on error
func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "not found")
}
