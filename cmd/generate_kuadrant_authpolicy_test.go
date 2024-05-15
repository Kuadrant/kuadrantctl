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
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"

	authorinoapi "github.com/kuadrant/authorino/api/v1beta2"
	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
)

var _ = Describe("Generate AuthPolicy", func() {
	var (
		cmd             *cobra.Command
		cmdStdoutBuffer *bytes.Buffer
		cmdStderrBuffer *bytes.Buffer
	)

	BeforeEach(func() {
		cmd = generateKuadrantAuthPolicyCommand()
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

	Context("with operation including security", func() {
		It("authorization policy generated", func() {
			cmd.SetArgs([]string{"--oas", "testdata/petstore_openapi.yaml"})
			Expect(cmd.Execute()).ShouldNot(HaveOccurred())
			out, err := ioutil.ReadAll(cmdStdoutBuffer)
			Expect(err).ShouldNot(HaveOccurred())

			var kap kuadrantapiv1beta2.AuthPolicy
			Expect(yaml.Unmarshal(out, &kap)).ShouldNot(HaveOccurred())
			Expect(kap.TypeMeta).To(Equal(metav1.TypeMeta{
				APIVersion: kuadrantapiv1beta2.GroupVersion.String(), Kind: "AuthPolicy",
			}))
			Expect(kap.ObjectMeta).To(Equal(metav1.ObjectMeta{
				Name:      "petstore",
				Namespace: "petstore-ns",
			}))
			Expect(kap.Spec.TargetRef).To(Equal(gatewayapiv1alpha2.PolicyTargetReference{
				Group:     gatewayapiv1.GroupName,
				Kind:      gatewayapiv1.Kind("HTTPRoute"),
				Name:      gatewayapiv1.ObjectName("petstore"),
				Namespace: ptr.To(gatewayapiv1.Namespace("petstore-ns")),
			}))
			Expect(kap.Spec.AuthPolicyCommonSpec.RouteSelectors).To(HaveExactElements(
				kuadrantapiv1beta2.RouteSelector{
					Matches: []gatewayapiv1.HTTPRouteMatch{
						{
							Path: &gatewayapiv1.HTTPPathMatch{
								Type:  ptr.To(gatewayapiv1.PathMatchExact),
								Value: ptr.To("/v1/dog"),
							},
							Method: ptr.To(gatewayapiv1.HTTPMethodPost),
						},
					},
				},
			))
			Expect(kap.Spec.AuthPolicyCommonSpec.AuthScheme).To(Equal(
				&kuadrantapiv1beta2.AuthSchemeSpec{
					Authentication: map[string]kuadrantapiv1beta2.AuthenticationSpec{
						"postDog_securedDog": kuadrantapiv1beta2.AuthenticationSpec{
							AuthenticationSpec: authorinoapi.AuthenticationSpec{
								Credentials: authorinoapi.Credentials{},
								AuthenticationMethodSpec: authorinoapi.AuthenticationMethodSpec{
									Jwt: &authorinoapi.JwtAuthenticationSpec{
										IssuerUrl: "https://example.com/.well-known/openid-configuration",
									},
								},
							},
							CommonAuthRuleSpec: kuadrantapiv1beta2.CommonAuthRuleSpec{
								RouteSelectors: []kuadrantapiv1beta2.RouteSelector{
									{
										Matches: []gatewayapiv1.HTTPRouteMatch{
											{
												Path: &gatewayapiv1.HTTPPathMatch{
													Type:  ptr.To(gatewayapiv1.PathMatchExact),
													Value: ptr.To("/v1/dog"),
												},
												Method: ptr.To(gatewayapiv1.HTTPMethodPost),
											},
										},
									},
								},
							},
						},
					},
				},
			))
		})
	})
})
