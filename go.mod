module github.com/pperiyasamy/irq-smp-balance

go 1.13

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/sirupsen/logrus v1.6.0
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/kubernetes v1.20.0-beta.0.0.20201030114605-f78d095d52a9
)

replace (
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20201029161059-917dcbabe107
	k8s.io/apiextensions-apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20201029161059-917dcbabe107
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20201029161059-917dcbabe107
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20201029161059-917dcbabe107
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20201029161059-917dcbabe107
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20201029161059-917dcbabe107
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20201029161059-917dcbabe107
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20201029161059-917dcbabe107
	k8s.io/code-generator => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20201029161059-917dcbabe107
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20201029161059-917dcbabe107
	k8s.io/component-helpers => k8s.io/kubernetes/staging/src/k8s.io/component-helpers v0.0.0-20201030114605-f78d095d52a9
	k8s.io/controller-manager => k8s.io/kubernetes/staging/src/k8s.io/controller-manager v0.0.0-20201029161059-917dcbabe107
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20201029161059-917dcbabe107
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20201029161059-917dcbabe107
	k8s.io/gengo => k8s.io/gengo v0.0.0-20200728071708-7794989d0000
	k8s.io/heapster => k8s.io/heapster v1.6.0-beta.1.0.20181130071115-e1e83412787b
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.3.1-0.20201028104956-52c62e3b70a9
	k8s.io/kube-aggregator => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20201029161059-917dcbabe107
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20201029161059-917dcbabe107
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200923155610-8b5066479488
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20201029161059-917dcbabe107
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20201029161059-917dcbabe107
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20201029161059-917dcbabe107
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20201029161059-917dcbabe107
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20201029161059-917dcbabe107
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20201029161059-917dcbabe107
	k8s.io/mount-utils => k8s.io/kubernetes/staging/src/k8s.io/mount-utils v0.0.0-20201030114605-f78d095d52a9
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20201029161059-917dcbabe107
	k8s.io/sample-cli-plugin => k8s.io/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20201029161059-917dcbabe107
	k8s.io/sample-controller => k8s.io/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20201029161059-917dcbabe107
	k8s.io/system-validators => k8s.io/system-validators v1.2.0
	k8s.io/utils => k8s.io/utils v0.0.0-20201027101359-01387209bb0d
)
