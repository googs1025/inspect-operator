package workload

import (
	"github.com/myoperator/inspectoperator/pkg/k8sconfig"
	"k8s.io/client-go/kubernetes"
)

var ClientSet kubernetes.Interface

func init() {
	ClientSet = k8sconfig.InitClient(k8sconfig.K8sRestConfig())
}
