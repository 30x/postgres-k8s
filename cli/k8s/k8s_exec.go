package k8s

import (
	"bytes"
	"io"
	"log"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	remotecommandserver "k8s.io/kubernetes/pkg/kubelet/server/remotecommand"
)

// GetK8sRestConfig returns a k8s rest client config
func GetK8sRestConfig() (conf *restclient.Config, err error) {
	// retrieve necessary kube config settings
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	return kubeConfig.ClientConfig()
}

// GetClient retrieves a kubernetes client
func GetClient() (*unversioned.Client, error) {
	// make a client config with kube config
	config, err := GetK8sRestConfig()
	if err != nil {
		return nil, err
	}

	// make a client out of the kube client config
	client, err := unversioned.New(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

//ExecCommand run the command in the namespace's pod and container TODO return stdout and stderr as streams.  Returns stdout, stderr, and error
func ExecCommand(namespace, podName, containerName string, commands []string) (io.ReadCloser, io.ReadCloser, error) {

	config, err := GetK8sRestConfig()

	if err != nil {
		return nil, nil, err
	}

	client, err := GetClient()

	if err != nil {
		return nil, nil, err
	}

	podExecOpts := &api.PodExecOptions{
		Container: containerName,
		Command:   commands,
		Stdin:     true, // redirect Stdin
		Stdout:    true, // reditect Stdout from container
		Stderr:    true, // redirect Stderr from container
		TTY:       true, // allocate a TTY
	}

	req := client.RESTClient.Post().
		Resource(api.ResourcePods.String()).
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	req.VersionedParams(podExecOpts, api.ParameterCodec) // not sure what this is, just followed examples/what kubectl does

	exec, err := remotecommand.NewExecutor(config, "POST", req.URL())

	if err != nil {
		return nil, nil, err
	}

	stdin := &bytes.Buffer{}

	// stdout := &bytes.Buffer{}
	// stderr := &bytes.Buffer{}

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	//todo, not sure we need this

	// io.WriteCloser

	log.Printf("About to execute remote stream")

	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdin:              stdin,
		Stdout:             stdoutWriter,
		Stderr:             stderrWriter,
		Tty:                false,
		TerminalSizeQueue:  nil,
	})

	log.Printf("Remote stream in progress")

	return stdoutReader, stderrReader, err

	// rmtCmd, err := remote.NewExecutor(restConf, "POST", req.URL())

	// if err != nil {
	// 	return err
	// }

	// supportedProtocols := []string{remoteServer.StreamProtocolV1Name, remoteServer.StreamProtocolV2Name}
	// err := rmtCmd.Stream(supportedProtocols, stdin, stdout, stderr, true)
}
