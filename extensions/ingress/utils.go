package ingress

import (
	"strconv"
	"strings"

	eirinix "github.com/SUSE/eirinix"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func getClientSet(manager eirinix.Manager) (clientset *kubernetes.Clientset, err error) {
	config, err := manager.GetKubeConnection()
	if err != nil {
		manager.GetLogger().Error((err.Error()))
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		manager.GetLogger().Error((err.Error()))
	}

	return
}

func getInstanceID(pod *corev1.Pod) string {
	instanceID := "0"
	el := strings.Split(pod.GetName(), "-")
	if len(el) != 0 {
		last := el[len(el)-1]
		if _, err := strconv.Atoi(last); err == nil {
			instanceID = last
		}
	}

	return instanceID
}
