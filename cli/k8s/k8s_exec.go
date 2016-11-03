package k8s

import (
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

//ExecCommand run the command in the namespace's pod and container TODO return stdout and stderr as streams.  Returns stdout, stderr, and error.  Note that this will block until streaming completes from the command
func ExecCommand(namespace, podName, containerName string, commands []string, stderr, stdout io.Writer) error {

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

	log.Printf("Requesting endpoint at %s", req.URL().String())

	exec, err := remotecommand.NewExecutor(config, "POST", req.URL())

	if err != nil {
		return err
	}

	// loggingWriter := &loggingWriter{
	// 	target: stdoutWriter,
	// }

	log.Printf("Remote stream in progress")

	err = exec.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandserver.SupportedStreamingProtocols,
		Stdin:              nil,
		Stdout:             stdout,
		Stderr:             stderr,
		Tty:                false,
		TerminalSizeQueue:  nil,
	})

	log.Printf("Remote stream closed")

	// return stdoutReader, stderrReader, err
	return err

}

type loggingWriter struct {
	target io.Writer
}

func (logger *loggingWriter) Write(p []byte) (n int, err error) {
	length := len(p)

	string := string(p)

	log.Printf("Writing %d bytes.  Bytes are '%s'", length, string)

	return logger.target.Write(p)
}
