package ingress

import (
	"encoding/json"
	"fmt"
	"strings"

	eirinix "github.com/SUSE/eirinix"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// AppNameAnnotation is the annotation label containing the Eirini application name
	AppNameAnnotation = "cloudfoundry.org/application_name"
	// AnnotationCopyKubernetesGenericLabels is the annotation which enables copy of
	// kubernetes generic labels that start with `app.kubernetes.org`
	AnnotationCopyKubernetesGenericLabels = "eirinix.suse.org/CopyKubeGenericLabels"
	// RoutesAnnotation is the annotation label containing the Eirini application routes
	RoutesAnnotation = "cloudfoundry.org/routes"
)

var (
	// KubeGenericLabelPrefix is the prefix of kubernetes generic label key
	KubeGenericLabelPrefix = "app.kubernetes.io"
)

// EiriniApp is the default RouteHandler for Eirini applications.
// It generates and reconciles the required data to handle routing with kubernetes native ingress
// resources.
type EiriniApp struct {
	GUID                        string
	Name                        string
	Namespace                   string
	PodName                     string
	InstanceID                  string
	Labels                      map[string]string
	Annotations                 map[string]string
	CopyKubernetesGenericLabels string
	Routes                      []Route
}

// Route represent a route information (hostname/port)
type Route struct {
	Hostname string
	Port     int
}

// NewEiriniApp returns a EiriniApp from a corev1.Pod
func NewEiriniApp(pod *corev1.Pod) (app EiriniApp) {
	var routes []Route

	app.GUID, _ = pod.GetLabels()[eirinix.LabelGUID]
	app.Name, _ = pod.GetAnnotations()[AppNameAnnotation] // we will use it for the service name
	app.CopyKubernetesGenericLabels = pod.GetAnnotations()[AnnotationCopyKubernetesGenericLabels]
	app.Labels = pod.GetLabels()
	app.Annotations = pod.GetAnnotations()

	app.Namespace = pod.GetNamespace()
	app.PodName = pod.GetName()
	app.InstanceID = getInstanceID(pod)
	routesJSON, _ := pod.GetAnnotations()[RoutesAnnotation] // [{"hostname":"dizzylizard.cap.xxxxx.nip.io","port":8080}]

	json.Unmarshal([]byte(routesJSON), &routes)
	app.Routes = routes
	return
}

// Validate returns true if we have enough information to handle routes
func (e EiriniApp) Validate() bool {
	return len(e.Routes) != 0 &&
		e.GUID != "" &&
		e.Name != "" &&
		e.Namespace != "" &&
		e.PodName != "" &&
		e.InstanceID != ""
}

// FirstInstance returns true if the pod is the first instance (e.g. if scaled or not)
func (e EiriniApp) FirstInstance() bool {
	return e.InstanceID == "0"
}

// UpdateService updates the given service from the Eirini app desired state
func (e EiriniApp) UpdateService(svc *corev1.Service, labels, annotations map[string]string) *corev1.Service {
	desired := e.DesiredService(labels, annotations)

	// Updates only the ports and meta
	svc.Annotations = desired.Annotations
	svc.Labels = desired.Labels
	svc.Spec.Ports = desired.Spec.Ports
	svc.Spec.Selector = desired.Spec.Selector
	return svc
}

// UpdateIngress updates the given ingress from the Eirini app desired state
func (e EiriniApp) UpdateIngress(in *v1beta1.Ingress, labels, annotations map[string]string, tls bool) *v1beta1.Ingress {
	desired := e.DesiredIngress(labels, annotations, tls)
	// Updates only the routes and meta
	in.Annotations = desired.Annotations
	in.Labels = desired.Labels
	in.Spec.Rules = desired.Spec.Rules
	in.Spec.TLS = desired.Spec.TLS

	return in
}

// DesiredService generates the desired service from the routes annotated in the Eirini App
func (e EiriniApp) DesiredService(labels, annotations map[string]string) *corev1.Service {
	ports := []corev1.ServicePort{}
	addedPorts := map[int]interface{}{}
	for _, route := range e.Routes {
		if _, ok := addedPorts[route.Port]; ok {
			continue
		}
		ports = append(ports, corev1.ServicePort{Port: int32(route.Port), TargetPort: intstr.FromInt(route.Port)})
		addedPorts[route.Port] = nil
	}

	if labels == nil {
		labels = map[string]string{}
	}

	// Copy kubernetes generic labels from pod to service
	if e.CopyKubernetesGenericLabels == "true" {
		for key, value := range e.Labels {
			if strings.Contains(key, KubeGenericLabelPrefix) {
				labels[key] = value
			}
		}
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.Name,
			Namespace:   e.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				eirinix.LabelGUID: e.GUID,
			},
		},
	}
}

// DesiredIngress generates the desired ingress from the routes annotated in the Eirini App
func (e EiriniApp) DesiredIngress(labels, annotations map[string]string, tls bool) *v1beta1.Ingress {
	rules := []v1beta1.IngressRule{}
	for _, route := range e.Routes {
		rules = append(rules, v1beta1.IngressRule{
			Host: route.Hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{{Path: "/",
						Backend: v1beta1.IngressBackend{
							ServiceName: e.DesiredService(labels, annotations).ObjectMeta.Name,
							ServicePort: intstr.FromInt(route.Port),
						},
					}},
				},
			},
		})
	}

	spec := v1beta1.IngressSpec{
		Rules: rules,
	}

	if tls {
		tlsEntry := []v1beta1.IngressTLS{}
		for _, route := range e.Routes {

			tlsEntry = append(tlsEntry,
				v1beta1.IngressTLS{
					Hosts:      []string{route.Hostname},
					SecretName: fmt.Sprintf("%s-tls", e.DesiredService(labels, annotations).ObjectMeta.Name),
				})
		}
		spec.TLS = tlsEntry
	}

	if labels == nil {
		labels = map[string]string{}
	}

	// Copy kubernetes generic labels from pod to ingress
	if e.CopyKubernetesGenericLabels == "true" {
		for key, value := range e.Labels {
			if strings.Contains(key, KubeGenericLabelPrefix) {
				labels[key] = value
			}
		}
	}

	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        e.Name,
			Namespace:   e.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: spec,
	}
}
