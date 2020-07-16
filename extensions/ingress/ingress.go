package ingress

import (
	"fmt"

	eirinix "github.com/SUSE/eirinix"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
)

type PodWatcher struct {
	GetRouteHandler func(*corev1.Pod) RouteHandler
}

func NewPodWatcher() *PodWatcher {
	return &PodWatcher{
		GetRouteHandler: func(pod *corev1.Pod) RouteHandler {
			return NewEiriniApp(pod)
		},
	}
}

func (pw *PodWatcher) Handle(manager eirinix.Manager, e watch.Event) {
	manager.GetLogger().Debug("Received event: ", e)

	if e.Object == nil {
		return
	}

	pod, ok := e.Object.(*corev1.Pod)
	if !ok {
		manager.GetLogger().Error("Received non-pod object in watcher channel")
		return
	}

	clientset, err := getClientSet(manager)
	if err != nil {
		manager.GetLogger().Error("Cannot generate clientset")
		return
	}

	app := pw.GetRouteHandler(pod)
	if !app.Validate() {
		fmt.Println("Missing app data", app)
		return
	}

	switch e.Type {
	case watch.Deleted:

		set := labels.Set(app.DesiredService().Spec.Selector)
		listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
		pods, err := clientset.CoreV1().Pods(pod.GetNamespace()).List(listOptions)
		if err != nil {
			manager.GetLogger().Error((err.Error()))
			return
		}
		// Don't delete if there are instances still running (scaling)
		if len(pods.Items) != 0 {
			return
		}

		err = clientset.CoreV1().Services(pod.GetNamespace()).Delete(app.DesiredService().GetName(), nil)
		if err != nil {
			manager.GetLogger().Error((err.Error()))
			//	return
		}
		fmt.Println("Deleted Services", app.DesiredService().GetName())

		err = clientset.ExtensionsV1beta1().Ingresses(pod.GetNamespace()).Delete(app.DesiredIngress().GetName(), nil)
		if err != nil {
			manager.GetLogger().Error((err.Error()))
			return
		}
		fmt.Println("Deleted ingress", app.DesiredIngress().GetName())

	default:
		if svc, err := clientset.CoreV1().Services(pod.GetNamespace()).Get(app.DesiredService().GetName(), metav1.GetOptions{}); err == nil {
			svc, err := clientset.CoreV1().Services(pod.GetNamespace()).Update(app.UpdateService(svc))
			if err != nil {
				manager.GetLogger().Error((err.Error()))
				//	return
			}
			fmt.Println("Updated service", svc.GetName())
		} else {
			svc, err := clientset.CoreV1().Services(pod.GetNamespace()).Create(app.DesiredService())
			if err != nil {
				manager.GetLogger().Error((err.Error()))
				//	return
			}
			fmt.Println("Created service", svc.GetName())
		}

		if ingr, err := clientset.ExtensionsV1beta1().Ingresses(pod.GetNamespace()).Get(app.DesiredIngress().GetName(), metav1.GetOptions{}); err == nil {
			ingr, err := clientset.ExtensionsV1beta1().Ingresses(pod.GetNamespace()).Update(app.UpdateIngress(ingr))
			if err != nil {
				manager.GetLogger().Error((err.Error()))
				//	return
			}
			fmt.Println("Updated Ingress", ingr.GetName())
		} else {
			ingr, err := clientset.ExtensionsV1beta1().Ingresses(pod.GetNamespace()).Create(app.DesiredIngress())
			if err != nil {
				manager.GetLogger().Error((err.Error()))
				return
			}
			fmt.Println("Created ingress", ingr.GetName())
		}

	}

	return
}
