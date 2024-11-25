package cmd

import (
	"bytes"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"

	kuadrantapiv1 "github.com/kuadrant/kuadrant-operator/api/v1"
)

var _ = Describe("Generate Ratelimitpolicy", func() {
	var (
		cmd             *cobra.Command
		cmdStdoutBuffer *bytes.Buffer
		cmdStderrBuffer *bytes.Buffer
	)

	BeforeEach(func() {
		cmd = generateKuadrantRateLimitPolicyCommand()
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

	Context("with rate limiting kuadrant extensions", func() {
		It("rate limit policy generated", func() {
			cmd.SetArgs([]string{"--oas", "testdata/petstore_openapi.yaml"})
			Expect(cmd.Execute()).ShouldNot(HaveOccurred())
			out, err := io.ReadAll(cmdStdoutBuffer)
			Expect(err).ShouldNot(HaveOccurred())

			var rlp kuadrantapiv1.RateLimitPolicy
			Expect(yaml.Unmarshal(out, &rlp)).ShouldNot(HaveOccurred())
			Expect(rlp.TypeMeta).To(Equal(metav1.TypeMeta{
				APIVersion: kuadrantapiv1.GroupVersion.String(), Kind: "RateLimitPolicy",
			}))
			Expect(rlp.ObjectMeta).To(Equal(metav1.ObjectMeta{
				Name:      "petstore",
				Namespace: "petstore-ns",
			}))
			Expect(rlp.Spec.TargetRef).To(Equal(gatewayapiv1alpha2.PolicyTargetReference{
				Group:     gatewayapiv1.GroupName,
				Kind:      gatewayapiv1.Kind("HTTPRoute"),
				Name:      gatewayapiv1.ObjectName("petstore"),
				Namespace: ptr.To(gatewayapiv1.Namespace("petstore-ns")),
			}))
			Expect(rlp.Spec.RateLimitPolicyCommonSpec.Limits).To(HaveLen(2))
			Expect(rlp.Spec.RateLimitPolicyCommonSpec.Limits).To(HaveKeyWithValue("getCat", kuadrantapiv1.Limit{
				Counters: []kuadrantapiv1.ContextSelector{
					"request.headers.x-forwarded-for",
				},
				RouteSelectors: []kuadrantapiv1.RouteSelector{
					{
						Matches: []gatewayapiv1.HTTPRouteMatch{
							{
								Path: &gatewayapiv1.HTTPPathMatch{
									Type:  ptr.To(gatewayapiv1.PathMatchExact),
									Value: ptr.To("/v1/cat"),
								},
								Method: ptr.To(gatewayapiv1.HTTPMethodGet),
							},
						},
					},
				},
				Rates: []kuadrantapiv1.Rate{
					{
						Limit:    1,
						Duration: 10,
						Unit:     kuadrantapiv1.TimeUnit("second"),
					},
				},
			}))
			Expect(rlp.Spec.RateLimitPolicyCommonSpec.Limits).To(HaveKeyWithValue("getDog", kuadrantapiv1.Limit{
				Counters: []kuadrantapiv1.ContextSelector{
					"request.headers.x-forwarded-for",
				},
				RouteSelectors: []kuadrantapiv1.RouteSelector{
					{
						Matches: []gatewayapiv1.HTTPRouteMatch{
							{
								Path: &gatewayapiv1.HTTPPathMatch{
									Type:  ptr.To(gatewayapiv1.PathMatchExact),
									Value: ptr.To("/v1/dog"),
								},
								Method: ptr.To(gatewayapiv1.HTTPMethodGet),
							},
						},
					},
				},
				Rates: []kuadrantapiv1.Rate{
					{
						Limit:    3,
						Duration: 10,
						Unit:     kuadrantapiv1.TimeUnit("second"),
					},
				},
			}))
		})
	})
})
