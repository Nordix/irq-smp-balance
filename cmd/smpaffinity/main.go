// Copyright (c) 2020-2021 Nordix Foundation.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)
	done := make(chan bool, 1)

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
	mutex := &sync.Mutex{}
	stopper := make(chan struct{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mutex.Lock()
			handleAddPod(obj.(*v1.Pod), cms)
			mutex.Unlock()
		},
		// don't need to handle update event.
		// resources (cpu) can't be edited.
		// editing IrqLabelSelector comes as add/delete event.
		DeleteFunc: func(obj interface{}) {
			mutex.Lock()
			handleDeletePod(obj.(*v1.Pod), cms)
			mutex.Unlock()
		},
	})

	var isRunning int32
	go func() {
		atomic.StoreInt32(&(isRunning), int32(1))
		// informer Run blocks until informer is stopped
		logrus.Infof("starting irq labeled pod informer")
		informer.Run(stopper)
		logrus.Infof("irq labeled pod informer is stopped")
		atomic.StoreInt32(&(isRunning), int32(0))
	}()

	go func() {
		sig := <-sigs
		logrus.Infof("received the signal %v", sig)
		done <- true
	}()
	// Capture signals to cleanup before exiting
	<-done

	close(stopper)
	tEnd := time.Now().Add(3 * time.Second)
	for tEnd.After(time.Now()) {
		if atomic.LoadInt32(&isRunning) == 0 {
			logrus.Infof("irq labeled pod informer is no longer running, proceed to force shutdown")
			break
		}
		time.Sleep(600 * time.Millisecond)
	}

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
		err = irq.SetIRQLoadBalancing(podCPUs, false, irq.IrqSmpAffinityProcFile, irq.PodIrqBannedCPUsFile)
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
		err = irq.SetIRQLoadBalancing(podCPUs, true, irq.IrqSmpAffinityProcFile, irq.PodIrqBannedCPUsFile)
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
