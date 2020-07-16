package ingress

import (
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
)

type RouteHandler interface {
	Validate() bool
	FirstInstance() bool
	UpdateService(svc *corev1.Service) *corev1.Service
	UpdateIngress(in *v1beta1.Ingress) *v1beta1.Ingress
	DesiredService() *corev1.Service
	DesiredIngress() *v1beta1.Ingress
}
