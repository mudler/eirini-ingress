package ingress

import (
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

type RouteHandler interface {
	Validate() bool
	FirstInstance() bool
	UpdateService(svc *corev1.Service, labels map[string]string) *corev1.Service
	UpdateIngress(in *v1beta1.Ingress, labels map[string]string) *v1beta1.Ingress
	DesiredService(map[string]string) *corev1.Service
	DesiredIngress(map[string]string) *v1beta1.Ingress
}
