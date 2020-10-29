package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang/glog"
	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
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
		glog.Errorf("no worker node name set in env variable")
		return
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf(" error with retrieving cluster config %v", err)
		return
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf(" error with configuring kube client %v", err)
		return
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		glog.Errorf(" error in listing down the pods %v", err)
		return
	}
	fmt.Printf("There are %d pods with resourceVersion %s in the cluster\n", len(pods.Items), pods.ListMeta.ResourceVersion)
	// qosClass is not a supported field selector
	watch, err := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{ResourceVersion: pods.ListMeta.ResourceVersion,
		LabelSelector: IrqLabelSelector,
		FieldSelector: fmt.Sprintf("spec.nodeName=%s,status.phase=Running,status", worker)})
	if err != nil {
		glog.Errorf(" error in listing down the pods %v", err)
		return
	}

	go func() {
		for event := range watch.ResultChan() {
			glog.Infof("Type: %v\n", event.Type)
			p, ok := event.Object.(*v1.Pod)
			if !ok {
				log.Fatal("error with pod event")
				continue
			}
			glog.Infof("%s, %s, %s, %s\n", p.ObjectMeta.Name, p.Status.Phase, p.Status.QOSClass, p.Spec.NodeName)
			//TODO:
		}
	}()

	// Capture signals to cleanup before exiting
	<-c
}
