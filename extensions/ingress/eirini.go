package ingress

import (
	"encoding/json"

	eirinix "github.com/SUSE/eirinix"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// AppNameAnnotation is the annotation label containing the Eirini application name
	AppNameAnnotation = "cloudfoundry.org/application_name"
	// RoutesAnnotation is the annotation label containing the Eirini application routes
	RoutesAnnotation = "cloudfoundry.org/routes"
)

// EiriniApp is the default RouteHandler for Eirini applications.
// It generates and reconciles the required data to handle routing with kubernetes native ingress
// resources.
type EiriniApp struct {
	GUID       string
	Name       string
	Namespace  string
	PodName    string
	InstanceID string
	Routes     []Route
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
func (e EiriniApp) UpdateService(svc *corev1.Service) *corev1.Service {
	desired := e.DesiredService()
	// Updates only the ports
	svc.Spec.Ports = desired.Spec.Ports
	svc.Spec.Selector = desired.Spec.Selector
	return svc
}

// UpdateIngress updates the given ingress from the Eirini app desired state
func (e EiriniApp) UpdateIngress(in *v1beta1.Ingress) *v1beta1.Ingress {
	desired := e.DesiredIngress()
	// Updates only the routes
	in.Spec.Rules = desired.Spec.Rules
	return in
}

// DesiredService generates the desired service from the routes annotated in the Eirini App
func (e EiriniApp) DesiredService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Name,
			Namespace: e.Namespace,
			Labels: map[string]string{
				"eirinix-ingress": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(e.Routes[0].Port)}},
			Selector: map[string]string{
				eirinix.LabelGUID: e.GUID,
			},
		},
	}
}

// DesiredIngress generates the desired ingress from the routes annotated in the Eirini App
func (e EiriniApp) DesiredIngress() *v1beta1.Ingress {
	rules := []v1beta1.IngressRule{}
	for _, route := range e.Routes {
		rules = append(rules, v1beta1.IngressRule{
			Host: route.Hostname,
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{{Path: "/",
						Backend: v1beta1.IngressBackend{
							ServiceName: e.DesiredService().ObjectMeta.Name,
							ServicePort: intstr.FromInt(route.Port),
						},
					}},
				},
			},
		})
	}

	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.Name,
			Namespace: e.Namespace,
			Labels: map[string]string{
				"k8s-app": "kube-controller-manager",
			},
		},
		Spec: v1beta1.IngressSpec{
			//	TLS: []v1beta1.IngressTLS{{Hosts: []string{}, SecretName: ""}},
			Rules: rules,
		},
	}
}
