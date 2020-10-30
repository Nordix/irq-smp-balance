package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("error with retrieving cluster config %v", err)
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("error with configuring kube client %v", err)
		return
	}

	startPODWatcher(clientset, worker)

	// Capture signals to cleanup before exiting
	<-c
}

func startPODWatcher(clientset *kubernetes.Clientset, worker string) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error in listing down the pods %v", err)
		return
	}
	// qosClass is not a supported field selector
	//TODO: watch for pod delete
	watch, err := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{ResourceVersion: pods.ListMeta.ResourceVersion,
		LabelSelector: IrqLabelSelector,
		FieldSelector: fmt.Sprintf("spec.nodeName=%s,status.phase=Running", worker)})
	if err != nil {
		logrus.Errorf(" error in listing down the pods %v", err)
		return
	}
	cms, err := irq.NewCPUManagerService()
	if err != nil {
		logrus.Errorf("error retrieving the cpumanager service")
		return
	}

	go func() {
		for event := range watch.ResultChan() {
			logrus.Infof("Type: %v\n", event.Type)
			pod, ok := event.Object.(*v1.Pod)
			if !ok {
				logrus.Errorf("error with pod event")
				continue
			}
			logrus.Infof("%s, %s, %s, %s\n", pod.ObjectMeta.Name, pod.Status.Phase, pod.Status.QOSClass, pod.Spec.NodeName)
			podCPUs := cms.GetAssignedCpus(string(pod.UID))
			if podCPUs != "" {
				//TODO:
				irq.SetIRQLoadBalancing(podCPUs, false, irq.IrqSmpAffinityProcFile)
			}
		}
	}()
}

