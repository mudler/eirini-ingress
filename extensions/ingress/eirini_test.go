package ingress_test

import (
	"fmt"

	eirinix "github.com/SUSE/eirinix"
	. "github.com/mudler/eirini-ingress/extensions/ingress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Route Handler", func() {
	Context("when evaluating pods", func() {

		var (
			app EiriniApp
		)

		BeforeEach(func() {
			app = NewEiriniApp(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "foo",
					Name:      "dizzylizard-test-79699025f0-0",
					Labels: map[string]string{
						eirinix.LabelGUID:        "test",
						"app.kubernetes.io/name": "foo",
					},
					Annotations: map[string]string{
						AnnotationCopyKubernetesGenericLabels: "true",
						AppNameAnnotation:                     "foo",
						RoutesAnnotation:                      `[{"hostname":"dizzylizard.cap.xxxxx.nip.io","port":8080}]`,
					},
				}})
		})

		Context("standard Eirini App", func() {
			It("decodes it correctly", func() {
				Expect(app.PodName).Should(Equal("dizzylizard-test-79699025f0-0"))
				Expect(app.Routes[0].Hostname).Should(Equal("dizzylizard.cap.xxxxx.nip.io"))
				Expect(app.CopyKubernetesGenericLabels).Should(Equal("true"))

				Expect(app.FirstInstance()).Should(BeTrue())
				Expect(app.Validate()).Should(BeTrue(), fmt.Sprint(app))

				Expect(len(app.DesiredService(nil, nil).Spec.Ports)).Should(Equal(1))
				Expect(app.DesiredService(nil, nil).Spec.Ports[0].TargetPort.String()).Should(Equal("8080"))
				Expect(app.DesiredService(nil, nil).Spec.Selector).Should(Equal(map[string]string{
					eirinix.LabelGUID: "test",
				}))

				Expect(len(app.DesiredIngress(nil, nil, false).Spec.Rules)).Should(Equal(1))
				Expect(app.DesiredIngress(nil, nil, false).Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName).Should(Equal(app.DesiredService(nil, nil).Name))
				Expect(app.DesiredIngress(nil, nil, false).Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.String()).Should(Equal(app.DesiredService(nil, nil).Spec.Ports[0].TargetPort.String()))
			})
		})

		Context("App updates", func() {
			var app2 EiriniApp
			BeforeEach(func() {
				app2 = NewEiriniApp(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "foo",
						Name:      "dizzylizard-test-79699025f0-0",
						Labels: map[string]string{
							eirinix.LabelGUID: "test2",
						},
						Annotations: map[string]string{
							AppNameAnnotation: "foo",
							RoutesAnnotation:  `[{"hostname":"dest.cap.xxxxx.nip.io","port":22}, {"hostname":"dizzylizard2.cap.xxxxx.nip.io","port":8080}]`,
						},
					}})
			})

			It("updates it correctly", func() {

				currentsvc := app.DesiredService(nil, nil)
				currentingr := app.DesiredIngress(nil, nil, true)
				app2.UpdateService(currentsvc, nil, nil)
				app2.UpdateIngress(currentingr, nil, nil, true)
				Expect(len(currentsvc.Spec.Ports)).Should(Equal(2))
				Expect(currentsvc.Spec.Ports[0].TargetPort.String()).Should(Equal("22"))
				Expect(currentsvc.Spec.Ports[1].TargetPort.String()).Should(Equal("8080"))
				Expect(currentsvc.Spec.Selector).Should(Equal(map[string]string{
					eirinix.LabelGUID: "test2",
				}))
				Expect(app2.Routes[0].Hostname).Should(Equal("dest.cap.xxxxx.nip.io"))

				Expect(len(currentingr.Spec.Rules)).Should(Equal(2))
				Expect(currentingr.Spec.Rules[0].HTTP.Paths[0].Backend.ServiceName).Should(Equal(app2.DesiredService(nil, nil).Name))
				Expect(currentingr.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.String()).Should(Equal(app2.DesiredService(nil, nil).Spec.Ports[0].TargetPort.String()))
				Expect(app2.Routes[1].Hostname).Should(Equal("dizzylizard2.cap.xxxxx.nip.io"))
				Expect(currentingr.Spec.Rules[1].HTTP.Paths[0].Backend.ServiceName).Should(Equal(app2.DesiredService(nil, nil).Name))
				Expect(currentingr.Spec.Rules[1].HTTP.Paths[0].Backend.ServicePort.String()).Should(Equal(app2.DesiredService(nil, nil).Spec.Ports[1].TargetPort.String()))
				Expect(len(currentingr.Spec.TLS)).To(Equal(2))
				Expect(currentingr.Spec.TLS[1].Hosts).Should(Equal([]string{"dizzylizard2.cap.xxxxx.nip.io"}))
				Expect(currentingr.Spec.TLS[1].SecretName).Should(Equal("foo-tls"))
				Expect(currentingr.Spec.TLS[0].Hosts).Should(Equal([]string{"dest.cap.xxxxx.nip.io"}))
				Expect(currentingr.Spec.TLS[0].SecretName).Should(Equal("foo-tls"))
			})
		})

		Context("Custom Annotations and Labels", func() {
			var app2 EiriniApp
			var testLabel, testAnnotations map[string]string
			BeforeEach(func() {
				app2 = NewEiriniApp(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "foo",
						Name:      "dizzylizard-test-79699025f0-0",
						Labels: map[string]string{
							"app.kubernetes.io/name": "foo",
						},
						Annotations: map[string]string{
							AnnotationCopyKubernetesGenericLabels: "true",
						},
					}})
				testLabel = map[string]string{"foo": "bar"}
				testAnnotations = map[string]string{"baz": "annotation"}
			})

			It("adds annotations and labels correctly", func() {
				currentsvc := app.DesiredService(testLabel, testAnnotations)
				currentingr := app.DesiredIngress(testLabel, testAnnotations, true)

				Expect(currentsvc.Annotations).Should(Equal(testAnnotations))
				Expect(currentsvc.Labels).Should(Equal(testLabel))
				Expect(currentingr.Annotations).Should(Equal(testAnnotations))
				Expect(currentingr.Labels).Should(Equal(testLabel))
			})

			It("updates annotations and labels correctly", func() {
				currentsvc := app.DesiredService(nil, nil)
				currentingr := app.DesiredIngress(nil, nil, true)
				app2.UpdateService(currentsvc, testLabel, testAnnotations)
				app2.UpdateIngress(currentingr, testLabel, testAnnotations, true)
				Expect(currentsvc.Annotations).Should(Equal(testAnnotations))
				Expect(currentsvc.Labels).Should(Equal(testLabel))
				Expect(currentingr.Annotations).Should(Equal(testAnnotations))
				Expect(currentingr.Labels).Should(Equal(testLabel))
			})
		})
	})
})
