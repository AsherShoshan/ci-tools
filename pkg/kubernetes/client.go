package kubernetes

import (
	"fmt"
	"time"

	coreapi "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/openshift/ci-tools/pkg/steps/loggingclient"
)

type PodClient interface {
	loggingclient.LoggingClient
	PendingTimeout() time.Duration
	// WithNewLoggingClient returns a new instance of the PodClient that resets
	// its LoggingClient.
	WithNewLoggingClient() PodClient
	Exec(namespace, pod string, opts *coreapi.PodExecOptions) (remotecommand.Executor, error)
	GetLogs(namespace, name string, opts *coreapi.PodLogOptions) *rest.Request
}

func NewPodClient(ctrlclient loggingclient.LoggingClient, config *rest.Config, client rest.Interface, pendingTimeout time.Duration) PodClient {
	return &podClient{
		LoggingClient:  ctrlclient,
		config:         config,
		client:         client,
		pendingTimeout: pendingTimeout,
	}
}

type podClient struct {
	loggingclient.LoggingClient
	config         *rest.Config
	client         rest.Interface
	pendingTimeout time.Duration
}

func (c podClient) PendingTimeout() time.Duration { return c.pendingTimeout }

func (c podClient) Exec(namespace, pod string, opts *coreapi.PodExecOptions) (remotecommand.Executor, error) {
	u := c.client.Post().Resource("pods").Namespace(namespace).Name(pod).SubResource("exec").VersionedParams(opts, scheme.ParameterCodec).URL()
	e, err := remotecommand.NewSPDYExecutor(c.config, "POST", u)
	if err != nil {
		return nil, fmt.Errorf("could not initialize a new SPDY executor: %w", err)
	}
	return e, nil
}

func (c podClient) GetLogs(namespace, name string, opts *coreapi.PodLogOptions) *rest.Request {
	return c.client.Get().Namespace(namespace).Name(name).Resource("pods").SubResource("log").VersionedParams(opts, scheme.ParameterCodec)
}

func (c podClient) WithNewLoggingClient() PodClient {
	c.LoggingClient = c.New()
	return c
}
