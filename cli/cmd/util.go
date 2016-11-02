package cmd

import (
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

//createClusterSelector create a selector for the cluster
func createClusterSelector(clusterName string) string {
	return fmt.Sprintf("app=postgres,cluster=%s", clusterName)
}

//getMasterPod get the master pod of the cluster
func getMasterPod(client *kubernetes.Clientset, namespace, clusterName string) (*v1.Pod, error) {
	selectorString := fmt.Sprintf("app=postgres,role=master,cluster=%s", clusterName)

	pods, err := client.Pods(namespace).List(v1.ListOptions{
		LabelSelector: selectorString,
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
	selectorString := fmt.Sprintf("app=postgres,role=replica,cluster=%s", clusterName)

	pods, err := client.Pods(namespace).List(v1.ListOptions{
		LabelSelector: selectorString,
	})

	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

//isNotFoundError Returns true if the resource is not found on error
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
