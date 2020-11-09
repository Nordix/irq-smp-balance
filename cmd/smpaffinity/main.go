package main

import (
	"fmt"
	"os"

	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	// WorkerNodeName env variable for worker name on which this code runs
	WorkerNodeName string = "WORKER_NODE_NAME"
	// IrqLabelSelector label selector for the pod which needs interrupt masking
	IrqLabelSelector string = "irq-load-balancing.docker.io=true"
)

func main() {

	c := irq.NewOSSignalChannel()

	worker, ok := os.LookupEnv(WorkerNodeName)
	if !ok {
		logrus.Errorf("no worker node name set in env variable")
		return
	}

	logrus.Infof("starting irq-smp-balance in %s", worker)

	// creates the in-cluster config
	clientSet := getClient()

	cms, err := irq.NewCPUManagerService()
	if err != nil {
		logrus.Errorf("error retrieving the cpumanager service")
		return
	}

	factory := informers.NewFilteredSharedInformerFactory(clientSet, 0, "", func(o *metav1.ListOptions) {
		o.LabelSelector = IrqLabelSelector
		o.FieldSelector = fmt.Sprintf("spec.nodeName=%s,status.phase=Running", worker)
	})
	informer := factory.Core().V1().Pods().Informer()
	stopper := make(chan struct{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleAddPod(obj.(*v1.Pod), cms)

		},
		DeleteFunc: func(obj interface{}) {
			handleDeletePod(obj.(*v1.Pod), cms)
		},
	})

	informer.Run(stopper)

	// Capture signals to cleanup before exiting
	<-c

	close(stopper)

	logrus.Infof("irq-smp-balance is stopped")
}

func handleAddPod(pod *v1.Pod, cms irq.CPUManagerService) {
	logrus.Infof("pod added %s, %s, %s, %s\n", pod.ObjectMeta.Name, pod.Status.Phase, pod.Status.QOSClass, pod.Spec.NodeName)
	if pod.Status.QOSClass != v1.PodQOSGuaranteed {
		logrus.Infof("pod %s is with %s qos class. ignoring", pod.ObjectMeta.Name, pod.Status.QOSClass)
		return
	}
	podUID := string(pod.UID)
	podCPUs, err := cms.GetAssignedCpus(podUID)
	if err != nil {
		logrus.Errorf("error in retrieving assigned cpus for pod %s: %v", pod.ObjectMeta.Name, err)
		return
	}
	logrus.Infof("assigned cpus %s for pod %s", podCPUs, pod.ObjectMeta.Name)
	if podCPUs != "" {
		err = irq.SetIRQLoadBalancing(podCPUs, false, irq.IrqSmpAffinityProcFile)
		if err != nil {
			logrus.Errorf("set irq load balancing for pod %s failed: %v", pod.ObjectMeta.Name, err)
			return
		}
	}
}

func handleDeletePod(pod *v1.Pod, cms irq.CPUManagerService) {
	logrus.Infof("pod deleted %s, %s, %s, %s\n", pod.ObjectMeta.Name, pod.Status.Phase, pod.Status.QOSClass, pod.Spec.NodeName)
	if pod.Status.QOSClass != v1.PodQOSGuaranteed {
		logrus.Infof("pod %s is with %s qos class. ignoring", pod.ObjectMeta.Name, pod.Status.QOSClass)
		return
	}
	podUID := string(pod.UID)
	podCPUs, err := cms.GetAssignedCpus(podUID)
	if err != nil {
		logrus.Warnf("not able to retrieve assigned cpus for pod %s: %v", pod.ObjectMeta.Name, err)
		podCPUs = cms.GetAssignedCpusFromCache(podUID)
	}
	logrus.Infof("assigned cpus %s for pod %s", podCPUs, pod.ObjectMeta.Name)
	if podCPUs != "" {
		err = irq.SetIRQLoadBalancing(podCPUs, true, irq.IrqSmpAffinityProcFile)
		if err != nil {
			logrus.Errorf("reset irq load balancing for pod %s failed: %v", pod.ObjectMeta.Name, err)
			return
		}
	}
	cms.Remove(podUID)
}

// GetClient returns a k8s clientset to the request from inside of cluster
func getClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("error with retrieving cluster config %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("error with configuring kube client %v", err)
	}

	return clientset
}
