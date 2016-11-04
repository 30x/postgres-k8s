package k8s

import (
	"bytes"
	"io"
	"sync"

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

//ExecCommandWithStdoutStderr Execute the command with a response of stdout, stderr, and error fully buffered nad parsed
func ExecCommandWithStdoutStderr(namespace, podName, containerName string, command []string) (string, string, error) {
	wg := &sync.WaitGroup{}

	wg.Add(3)

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	stdoutBuff := &bytes.Buffer{}
	stderrBuff := &bytes.Buffer{}

	var execError error
	var stdErrCopyErr error
	var stdOutCopyErr error

	//start the stderr in the background
	go func() {
		defer wg.Done()
		// log.Printf("Starting stderr read")
		_, stdErrCopyErr = io.Copy(stdoutBuff, stderrReader)
		// log.Printf("Completed stderr read")
	}()

	//start the stdout in the background
	go func() {
		defer wg.Done()
		// log.Printf("Starting stdout read")
		_, stdOutCopyErr = io.Copy(stderrBuff, stdoutReader)
		// log.Printf("Completed stdout read")
	}()

	//start the command in the background
	go func() {
		defer wg.Done()
		// log.Printf("Executing the command")
		execError = ExecCommand(namespace, podName, containerName, command, stdoutWriter, stderrWriter)
	}()

	// log.Printf("Waiting for stdout and stderr to complete streaming.")
	//wait for all the copying to complete
	wg.Wait()

	// log.Printf("Returning strings and errs")

	stdout := stdoutBuff.String()
	stderr := stderrBuff.String()

	if execError != nil {
		return "", "", execError
	}

	if stdErrCopyErr != nil {
		return "", "", stdErrCopyErr
	}

	if stdOutCopyErr != nil {
		return "", "", stdOutCopyErr
	}

	return stdout, stderr, nil

}

//ExecCommand run the command in the namespace's pod and container TODO return stdout and stderr as streams.  Returns stdout, stderr, and error.  Note that this will block until streaming completes from the command
func ExecCommand(namespace, podName, containerName string, commands []string, stderr, stdout *io.PipeWriter) error {

	config, err := GetK8sRestConfig()

	if err != nil {
		return err
	}

	client, err := GetClient()

	if err != nil {
		return err
	}

	podExecOpts := &api.PodExecOptions{
		Container: containerName,
		Command:   commands,
		Stdin:     false, // redirect Stdin
		Stdout:    true,  // reditect Stdout from container
		Stderr:    true,  // redirect Stderr from container
		TTY:       false, // allocate a TTY
	}

	req := client.RESTClient.Post().
		Resource(api.ResourcePods.String()).
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", containerName)

	req.VersionedParams(podExecOpts, api.ParameterCodec) // not sure what this is, just followed examples/what kubectl does

	// log.Printf("Requesting endpoint at %s", req.URL().String())

	exec, err := remotecommand.NewExecutor(config, "POST", req.URL())

	if err != nil {
		return err
	}

	// log.Printf("Remote stream in progress")

	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdin:              nil,
		Stdout:             stdout,
		Stderr:             stderr,
		Tty:                false,
		TerminalSizeQueue:  nil,
	})

	// log.Printf("Remote stream closed")

	stderr.Close()
	stdout.Close()

	// return stdoutReader, stderrReader, err
	return err

}
