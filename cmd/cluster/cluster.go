package cluster

import (
	"context"
	"fmt"
	"os"

	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

type Cluster struct {
	client        *corev1client.CoreV1Client
	appsClient    *appsv1client.AppsV1Client
	metricsClient *metrics.Clientset
	replicas      map[string]int32
}

func NewCluster(appConfig *model.Config) *Cluster {
	if appConfig.WatchDog.Namespace == "" &&
		len(appConfig.WatchDog.Actions) == 0 {
		log.Info("Missing watchdog configuration. Cluster configuration ignored.")
		return nil
	}

	var config *rest.Config
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); os.IsNotExist(err) {
		// Instantiate loader for kubeconfig file.
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		// Get a rest.Config from the kubeconfig file.  This will be passed into all
		// the client objects we create.
		config, err = kubeconfig.ClientConfig()
		if err != nil {
			log.Error(fmt.Sprintf("Error creating client connection: %s", err.Error()))
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Error(fmt.Sprintf("Error creating client connection: %s", err.Error()))
		}
	}

	// Create a Kubernetes core/v1 client.
	coreClient, err := corev1client.NewForConfig(config)
	if err != nil {
		log.Error(fmt.Sprintf("Error creating client: %s", err.Error()))
	}

	appsClient, err := appsv1client.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	wd := Cluster{
		client:        coreClient,
		appsClient:    appsClient,
		metricsClient: metricsClient,
		replicas:      make(map[string]int32),
	}

	return &wd
}

// Scale down each deployment in namespace
func (wd *Cluster) ScaleDown(names []string, namespace string) (err error) {
	for i := range names {
		err = wd.scaleDown(names[i], namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wd *Cluster) scaleDown(name string, namespace string) error {
	log.Info(fmt.Sprintf("scale down %s in %s", name, namespace))

	// get current scale
	specs, err := wd.appsClient.Deployments(namespace).
		GetScale(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("error while get specs from deployment %s: %s", name, err.Error()))
		return err
	}

	// remember scale count
	if specs.Spec.Replicas > 0 {
		wd.replicas[name] = specs.Spec.Replicas
	}
	specs.Spec.Replicas = 0

	// set scale count
	_, err = wd.appsClient.Deployments(namespace).
		UpdateScale(context.TODO(), name, specs, metav1.UpdateOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("error while set specs from deployment %s: %s", name, err.Error()))
	}

	//todo wait until down

	return err
}

// Scale up each deployment in namespace
func (wd *Cluster) ScaleUp(names []string, namespace string) (err error) {
	for i := range names {
		err = wd.scaleUp(names[i], namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func (wd *Cluster) scaleUp(name string, namespace string) error {
	log.Info(fmt.Sprintf("scale up %s in %s", name, namespace))

	// get deployment specs
	specs, err := wd.appsClient.Deployments(namespace).
		GetScale(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("error while get specs from deployment %s: %s", name, err.Error()))
		return err
	}

	// restore scale count
	if wd.replicas[name] == 0 {
		wd.replicas[name] = 1
	}
	specs.Spec.Replicas = wd.replicas[name]

	// scale up deployment
	_, err = wd.appsClient.Deployments(namespace).
		UpdateScale(context.TODO(), name, specs, metav1.UpdateOptions{})
	if err != nil {
		log.Error(fmt.Sprintf("error while set specs from deployment %s: %s", name, err.Error()))
		return err
	}

	return nil
}

func (wd *Cluster) DeletePod(name string, namespace string) error {
	log.Info(fmt.Sprintf("killing %s in %s", name, namespace))

	// List all Pods in our current Namespace.
	pods, err := wd.client.Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})
	if err != nil {
		log.Error(fmt.Sprintf("Error while list all pods: %s", err.Error()))
	}

	log.Info(fmt.Sprintf("Pods to delete in namespace %s:", namespace))
	for i := range pods.Items {
		err = wd.client.Pods(namespace).Delete(context.Background(), pods.Items[i].Name, metav1.DeleteOptions{})
		if err != nil {
			log.Error(fmt.Sprintf("Error while delete pod %s: %s", pods.Items[i].Name, err.Error()))
		} else {
			log.Info(fmt.Sprintf("Pod %s deleted", pods.Items[i].Name))
		}
	}

	return nil
}

func (wd *Cluster) GetPodIp(name string, namespace string) ([]string, error) {
	log.Info(fmt.Sprintf("Pods to delete in namespace %s:", namespace))
	pods, err := wd.client.Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})
	if err != nil {
		log.Error(fmt.Sprintf("Error while get pod %s: %s", name, err.Error()))
	}

	result := make([]string, 0)
	for i := range pods.Items {
		result = append(result, pods.Items[i].Status.PodIP)
	}

	return result, nil
}

func (wd *Cluster) GetPodMemory(name string, namespace string) ([]int64, error) {
	options := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	}
	podMetrics, err := wd.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), options)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}

	result := make([]int64, 0)
	for _, podMetric := range podMetrics.Items {
		podContainers := podMetric.Containers
		for _, container := range podContainers {
			// cpuQuantity, ok := container.Usage.Cpu().AsInt64()
			// if !ok {
			// 	//return nil, nil
			// }
			memQuantity, ok := container.Usage.Memory().AsInt64()
			if !ok {
				log.Error(fmt.Sprintf("error while load memory from pod %s", container.Name))
			} else {
				result = append(result, memQuantity)
			}
			// msg := fmt.Sprintf("Container Name: %s \n CPU usage: %d \n Memory usage: %d", container.Name, cpuQuantity, memQuantity)
			// fmt.Println(msg)
		}

	}

	return result, nil
}

func (wd *Cluster) Test() error {
	//var kubeconfig *string
	//if home := homeDir(); home != "" {
	//	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()

	//// use the current context in kubeconfig
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	return err
	//}

	//buildV1Client, err := buildv1.NewForConfig(config)
	//if err != nil {
	//	return err
	//}

	namespace := ""

	//// get all builds
	//builds, err := buildV1Client.Builds(namespace).List(context.TODO(), metav1.ListOptions{})
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("There are %d builds in project %s\n", len(builds.Items), namespace)
	//// List names of all builds
	//for i, build := range builds.Items {
	//	fmt.Printf("index %d: Name of the build: %s", i, build.Name)
	//}
	//
	//// get a specific build
	//build := "cakephp-ex-1"
	//myBuild, err := buildV1Client.Builds(namespace).Get(context.TODO(), build, metav1.GetOptions{})
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("Found build %s in namespace %s\n", build, namespace)
	//fmt.Printf("Raw printout of the build %+v\n", myBuild)
	//// get details of the build
	//fmt.Printf("name %s, start time %s, duration (in sec) %.0f, and phase %s\n",
	//	myBuild.Name, myBuild.Status.StartTimestamp.String(),
	//	myBuild.Status.Duration.Seconds(), myBuild.Status.Phase)
	//
	//// trigger a build
	//buildConfig := "cakephp-ex"
	//myBuildConfig, err := buildV1Client.BuildConfigs(namespace).Get(context.TODO(), buildConfig, metav1.GetOptions{})
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("Found BuildConfig %s in namespace %s\n", myBuildConfig.Name, namespace)
	//buildRequest := v1.BuildRequest{}
	//buildRequest.Kind = "BuildRequest"
	//buildRequest.APIVersion = "build.openshift.io/v1"
	//objectMeta := metav1.ObjectMeta{}
	//objectMeta.Name = "cakephp-ex"
	//buildRequest.ObjectMeta = objectMeta
	//buildTriggerCause := v1.BuildTriggerCause{}
	//buildTriggerCause.Message = "Manually triggered"
	//buildRequest.TriggeredBy = []v1.BuildTriggerCause{buildTriggerCause}
	//myBuild, err = buildV1Client.BuildConfigs(namespace).Instantiate(context.TODO(), objectMeta.Name, &buildRequest, metav1.CreateOptions{})
	//
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("Name of the triggered build %s\n", myBuild.Name)

	// Instantiate loader for kubeconfig file.
	kubeconfig1 := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restconfig, err := kubeconfig1.ClientConfig()
	if err != nil {
		panic(err)
	}

	// Create a Kubernetes core/v1 client.
	coreclient, err := corev1client.NewForConfig(restconfig)
	if err != nil {
		panic(err)
	}

	// appsClient, err := appsv1client.NewForConfig(restconfig)
	// if err != nil {
	// 	panic(err)
	// }

	// List all Pods in our current Namespace.
	pods, err := coreclient.Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("Pods in namespace %s:", namespace))
	for i := range pods.Items {
		log.Info(fmt.Sprintf("  %s", pods.Items[i].Name))
	}

	// List all Pods in our current Namespace.
	pods1, err := coreclient.Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: "app=json-server",
	})
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("Pods in namespace %s:", namespace))
	for _, pod := range pods1.Items {
		log.Info(fmt.Sprintf("  %s", pod.Name))
	}

	//err = coreclient.Pods(namespace).Delete(context.Background(), "json-server-57bbd69859-bcshr", metav1.DeleteOptions{})
	//if err != nil {
	//	panic(err)
	//}

	return nil
}
