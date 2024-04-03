package clients

import (
	"fmt"
	"os"

	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubernetesClient struct {
	client        *corev1client.CoreV1Client
	appsClient    *appsv1client.AppsV1Client
	metricsClient *metrics.Clientset
	config *model.Config
}

func NewKubernetesClient(config *model.Config) (*KubernetesClient, error) {
	var restConfig *rest.Config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); os.IsNotExist(err) {
		// Instantiate loader for kubeconfig file.
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		// Get a rest.Config from the kubeconfig file.  This will be passed into all
		// the client objects we create.
		restConfig, err = kubeconfig.ClientConfig()
		if err != nil {
			log.Error(fmt.Sprintf("Error creating client connection: %s", err.Error()))
			return nil, err
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Error(fmt.Sprintf("Error creating client connection: %s", err.Error()))
			return nil, err
		}
	}

	// Create a Kubernetes core/v1 client.
	coreClient, err := corev1client.NewForConfig(restConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Error creating client: %s", err.Error()))
		return nil, err
	}

	appsClient, err := appsv1client.NewForConfig(restConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Error creating client: %s", err.Error()))
		return nil, err
	}

	metricsClient, err := metrics.NewForConfig(restConfig)
	if err != nil {
		log.Error(fmt.Sprintf("Error creating client: %s", err.Error()))
		return nil, err
	}

	kc := KubernetesClient {
		client:        coreClient,
		appsClient:    appsClient,
		metricsClient: metricsClient,
		config: config,
	}

	return &kc, nil
}

func (kc *KubernetesClient) Execute(job *model.Job, channel chan *model.TaskResult) {
	switch job.Type {
	case "memory":
		result, err := kc.GetPodMemory(job.Label, job.Namespace)
		if err != nil {
			log.Error(fmt.Sprintf("error while scaling down %s: %s", job.Label, err.Error()))
		}

		r := model.TaskResult{
			Duration: result[0],
		}
		
		channel <- &r
	}
}
