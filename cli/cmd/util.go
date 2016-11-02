package cmd

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

//createClusterSelector create a selector for the cluster
func createClusterSelector(clusterName string) string {
	return fmt.Sprintf("app=postgres,cluster=%s", clusterName)
}

//getMasterPod get the master pod of the cluster
func getMasterPod(client *kubernetes.Clientset, namespace, clusterName string) (*v1.Pod, error) {
	selectorString := fmt.Sprintf("app=postgres,type=write,cluster=%s", clusterName)

	pods, err := client.Pods(namespace).List(v1.ListOptions{
		LabelSelector: selectorString,
	})

	if err != nil {
		return nil, err
	}

	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("Could not find master node in cluster %s", clusterName)
	}

	return &pods.Items[0], nil
}

//getMasterPod get the master pod of the cluster
func getReplicaPods(client *kubernetes.Clientset, namespace, clusterName string) ([]v1.Pod, error) {
	selectorString := fmt.Sprintf("app=postgres,type=read,cluster=%s", clusterName)

	pods, err := client.Pods(namespace).List(v1.ListOptions{
		LabelSelector: selectorString,
	})

	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}
