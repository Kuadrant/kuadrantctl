package cmd

import (
	"bytes"
	"io/ioutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Generate HTTPRoute", func() {
	var (
		cmd             *cobra.Command
		cmdStdoutBuffer *bytes.Buffer
		cmdStderrBuffer *bytes.Buffer
	)

	BeforeEach(func() {
		cmd = generateGatewayApiHttpRouteCommand()
		cmdStdoutBuffer = bytes.NewBufferString("")
		cmdStderrBuffer = bytes.NewBufferString("")
		cmd.SetOut(cmdStdoutBuffer)
		cmd.SetErr(cmdStderrBuffer)
	})

	Context("with invalid OAS", func() {
		It("happy path", func() {
			cmd.SetArgs([]string{"--oas", "testdata/invalid_oas.yaml"})
			Expect(cmd.Execute()).Should(MatchError(ContainSubstring("OpenAPI validation error")))

		})
	})

	Context("with root level kuadrant extensions", func() {
		It("HTTPRoute is generated", func() {
			cmd.SetArgs([]string{"--oas", "testdata/petstore_openapi.yaml"})
			Expect(cmd.Execute()).ShouldNot(HaveOccurred())
			out, err := ioutil.ReadAll(cmdStdoutBuffer)
			Expect(err).ShouldNot(HaveOccurred())

			var httpRoute gatewayapiv1.HTTPRoute
			Expect(yaml.Unmarshal(out, &httpRoute)).ShouldNot(HaveOccurred())
			Expect(httpRoute.TypeMeta).To(Equal(metav1.TypeMeta{
				APIVersion: gatewayapiv1.GroupVersion.String(),
				Kind:       "HTTPRoute",
			}))
			Expect(httpRoute.ObjectMeta).To(Equal(metav1.ObjectMeta{
				Name:      "petstore",
				Namespace: "petstore-ns",
			}))
			Expect(httpRoute.Spec.CommonRouteSpec).To(Equal(gatewayapiv1.CommonRouteSpec{
				ParentRefs: []gatewayapiv1.ParentReference{
					{
						Name: "gw", Namespace: ptr.To(gatewayapiv1.Namespace("gw-ns")),
					},
				},
			}))
			Expect(httpRoute.Spec.Hostnames).To(Equal([]gatewayapiv1.Hostname{
				gatewayapiv1.Hostname("example.com"),
			}))
			Expect(httpRoute.Spec.Rules).To(HaveLen(3))
			Expect(httpRoute.Spec.Rules).To(ContainElement(
				gatewayapiv1.HTTPRouteRule{
					Matches: []gatewayapiv1.HTTPRouteMatch{
						{
							Path: &gatewayapiv1.HTTPPathMatch{
								Type:  ptr.To(gatewayapiv1.PathMatchExact),
								Value: ptr.To("/v1/cat"),
							},
							Method: ptr.To(gatewayapiv1.HTTPMethodGet),
						},
					},
					BackendRefs: []gatewayapiv1.HTTPBackendRef{
						{
							BackendRef: gatewayapiv1.BackendRef{
								BackendObjectReference: gatewayapiv1.BackendObjectReference{
									Name:      "petstore",
									Namespace: ptr.To(gatewayapiv1.Namespace("petstore")),
									Port:      ptr.To(gatewayapiv1.PortNumber(80)),
								},
							},
						},
					},
				},
			))
			Expect(httpRoute.Spec.Rules).To(ContainElement(
				gatewayapiv1.HTTPRouteRule{
					Matches: []gatewayapiv1.HTTPRouteMatch{
						{
							Path: &gatewayapiv1.HTTPPathMatch{
								Type:  ptr.To(gatewayapiv1.PathMatchExact),
								Value: ptr.To("/v1/dog"),
							},
							Method: ptr.To(gatewayapiv1.HTTPMethodGet),
						},
					},
					BackendRefs: []gatewayapiv1.HTTPBackendRef{
						{
							BackendRef: gatewayapiv1.BackendRef{
								BackendObjectReference: gatewayapiv1.BackendObjectReference{
									Name:      "petstore",
									Namespace: ptr.To(gatewayapiv1.Namespace("petstore")),
									Port:      ptr.To(gatewayapiv1.PortNumber(80)),
								},
							},
						},
					},
				},
			))
			Expect(httpRoute.Spec.Rules).To(ContainElement(
				gatewayapiv1.HTTPRouteRule{
					Matches: []gatewayapiv1.HTTPRouteMatch{
						{
							Path: &gatewayapiv1.HTTPPathMatch{
								Type:  ptr.To(gatewayapiv1.PathMatchExact),
								Value: ptr.To("/v1/dog"),
							},
							Method: ptr.To(gatewayapiv1.HTTPMethodPost),
						},
					},
					BackendRefs: []gatewayapiv1.HTTPBackendRef{
						{
							BackendRef: gatewayapiv1.BackendRef{
								BackendObjectReference: gatewayapiv1.BackendObjectReference{
									Name:      "petstore",
									Namespace: ptr.To(gatewayapiv1.Namespace("petstore")),
									Port:      ptr.To(gatewayapiv1.PortNumber(80)),
								},
							},
						},
					},
				},
			))
		})
	})
})
