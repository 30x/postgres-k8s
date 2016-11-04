package cmd

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/30x/postgres-k8s/cli/k8s"

	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	extv1beta1 "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.4/pkg/labels"
)

//createClusterSelector create a selector for the cluster
func createClusterSelector(clusterName string) string {
	return fmt.Sprintf("app=postgres,cluster=%s", clusterName)
}

//getMasterPod get the master pod of the cluster
func getMasterPod(client *kubernetes.Clientset, namespace, clusterName string) (*v1.Pod, error) {
	selectorString := fmt.Sprintf("app=postgres,master=true,cluster=%s", clusterName)

	selector, err := labels.Parse(selectorString)

	if err != nil {
		return nil, err
	}

	pods, err := client.Pods(namespace).List(api.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("Master node not found in cluster %s", clusterName)
	}

	if len(pods.Items) > 1 {
		masters := ""

		for idx, pod := range pods.Items {
			if idx == 0 {
				masters = pod.Name
			} else {
				masters = masters + "," + pod.Name
			}
		}
		return nil, fmt.Errorf("More than 1 master detected. Pods with names %s were found.  Cannot determine which is master.  Remove the replica set controlling the old master", masters)
	}

	return &pods.Items[0], nil
}

//getMasterPod get the master pod of the cluster
func getReplicaPods(client *kubernetes.Clientset, namespace, clusterName string) ([]v1.Pod, error) {
	selectorString := fmt.Sprintf("app=postgres,cluster=%s", clusterName)

	selector, err := labels.Parse(selectorString)

	if err != nil {
		return nil, err
	}

	pods, err := client.Pods(namespace).List(api.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	//now remove the master
	podsToReturn := []v1.Pod{}

	for _, pod := range pods.Items {
		if pod.Labels["master"] != "true" {
			podsToReturn = append(podsToReturn, pod)
		}
	}

	return podsToReturn, nil
}

//getClusterReplicaSets get the replica sets for this cluster
func getClusterReplicaSets(client *kubernetes.Clientset, namespace, clusterName string) ([]extv1beta1.ReplicaSet, error) {
	selectorString := fmt.Sprintf("app=postgres,cluster=%s", clusterName)

	selector, err := labels.Parse(selectorString)

	if err != nil {
		return nil, err
	}

	replicaSets, err := client.ReplicaSets(namespace).List(api.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	return replicaSets.Items, nil
}

//getPodsForReplicaSet get the pods for a replica set
func getPodsForReplicaSet(client *kubernetes.Clientset, rs *extv1beta1.ReplicaSet) ([]v1.Pod, error) {

	rsLabels := rs.Spec.Template.GetLabels()

	//TODO, this is hacky, but it's not clear how to easily and cleanly convert from a set of labels in the resposne into a selector without reparsing
	selectorStringBuff := &bytes.Buffer{}

	for key, value := range rsLabels {

		if selectorStringBuff.Len() != 0 {
			selectorStringBuff.WriteString(",")
		}

		selectorStringBuff.WriteString(key)
		selectorStringBuff.WriteString("=")
		selectorStringBuff.WriteString(value)

	}

	selector, err := labels.Parse(selectorStringBuff.String())

	if err != nil {
		return nil, err
	}

	pods, err := client.Pods(rs.Namespace).List(api.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

//getActiveIndexesForCluster get all indexes currently used in the cluster.  They are returned sorted from low to high
func getActiveIndexesForCluster(pods []v1.Pod) ([]int, error) {
	indexes := []int{}

	for _, pod := range pods {

		index, err := getPodIndex(&pod)

		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)

	}

	sort.Ints(indexes)

	return indexes, nil
}

//updateMasterConfigurationWithReplicaIds set the master configuration with the specified replica ids
func updateMasterConfigurationWithReplicaIds(namespace, podName, containerName string, indexesToSet []int) error {
	//now convert them to a string
	replicaNames := ""

	for index, replicaIndex := range indexesToSet {
		indexAsString := strconv.Itoa(replicaIndex)
		if index == 0 {
			replicaNames = indexAsString
		} else {
			replicaNames = replicaNames + "," + indexAsString
		}
	}

	command := []string{"bash", "/clusterutils/setreplicas.sh", replicaNames}

	_, stderr, err := k8s.ExecCommandWithStdoutStderr(namespace, podName, containerName, command)

	if err != nil {
		return err
	}

	if len(stderr) != 0 {
		return fmt.Errorf("An error occured when running the command set replicas command on postgres.  The stderr output is below.  \n %s", stderr)
	}

	return nil

}

//getPodIndex get the index of the pod
func getPodIndex(pod *v1.Pod) (int, error) {
	indexLabel, ok := pod.Labels["index"]

	if !ok {
		return 0, fmt.Errorf("Could not find 'index' label on pod %s", pod.Name)
	}

	index, err := strconv.Atoi(indexLabel)

	if err != nil {
		return 0, err
	}

	return index, nil
}

//isNotFoundError Returns true if the resource is not found on error
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

//testMaster Tests the master pod is functioning correctly by executing a create database, insert, query, then delete dataabase
func testMaster(client *kubernetes.Clientset, namespace, clusterName string) error {

	masterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil {
		return err
	}

	log.Printf("Validating postgres is functioning properly.  This may take a bit as the service normalizes.")

	command := []string{"bash", "/clusterutils/testdb.sh"}
	// command := []string{"echo", "Marco"}

	if masterPod.Status.Phase != v1.PodRunning {
		return fmt.Errorf("cannot exec into a container in a non running pod; current phase is %s", masterPod.Status.Phase)
	}

	containerName, err := getContainerNameForPod(masterPod)

	if err != nil {
		return err
	}

	stdout, stderr, err := k8s.ExecCommandWithStdoutStderr(namespace, masterPod.Name, containerName, command)

	if err != nil {
		return err
	}

	if !strings.Contains(stdout, expectedQueryOutput) {
		return fmt.Errorf("Could not find the string '%s' in the output. Validate that postgres has started manually", expectedQueryOutput)
	}

	if len(stderr) != 0 {
		return fmt.Errorf("An error occured when running the command on postgres.  The stderr output is below.  \n %s", stderr)
	}

	return nil

}

//Downloads the index information from replicas, then inserts into the master pod
func updateMasterWithReplicas(client *kubernetes.Clientset, namespace, clusterName string) error {
	//now update the master with the new ids

	masterPod, err := getMasterPod(client, namespace, clusterName)

	if err != nil {
		return err
	}

	pods, err := getReplicaPods(client, namespace, clusterName)

	if err != nil {
		return err
	}

	replicaIndexes, err := getActiveIndexesForCluster(pods)

	if err != nil {
		return err
	}

	containerName, err := getContainerNameForPod(masterPod)

	if err != nil {
		return err
	}

	err = updateMasterConfigurationWithReplicaIds(namespace, masterPod.Name, containerName, replicaIndexes)

	return err
}

func getContainerNameForPod(pod *v1.Pod) (string, error) {

	if len(pod.Spec.Containers) > 1 {
		return "", fmt.Errorf("Only 1 container per pod is supported")
	}

	return pod.Spec.Containers[0].Name, nil
}
