package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"

	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/kuadrantapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

//kuadrantctl generate kuadrant ratelimitpolicy --oas [OAS_FILE_PATH | OAS_URL | @]

var (
	generateRateLimitPolicyOAS    string
	generateRateLimitPolicyFormat string
)

func generateKuadrantRateLimitPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ratelimitpolicy",
		Short: "Generate Kuadrant Rate Limit Policy from OpenAPI 3.0.X",
		Long:  "Generate Kuadrant Rate Limit Policy from OpenAPI 3.0.X",
		RunE:  runGenerateKuadrantRateLimitPolicy,
	}

	cmd.Flags().StringVar(&generateRateLimitPolicyOAS, "oas", "", "Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required)")
	cmd.Flags().StringVarP(&generateRateLimitPolicyFormat, "output-format", "o", "yaml", "Output format: 'yaml' or 'json'.")

	if err := cmd.MarkFlagRequired("oas"); err != nil {
		fmt.Println("Error setting 'oas' flag as required:", err)
		os.Exit(1)
	}

	return cmd
}

func runGenerateKuadrantRateLimitPolicy(cmd *cobra.Command, args []string) error {
	oasDataRaw, err := utils.ReadExternalResource(generateRateLimitPolicyOAS)
	if err != nil {
		return err
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData(oasDataRaw)
	if err != nil {
		return err
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		return fmt.Errorf("OpenAPI validation error: %w", err)
	}

	rlp := buildRateLimitPolicy(doc)

	jsonBytes, err := json.Marshal(rlp)
	if err != nil {
		return err
	}

	var outputBytes []byte
	if generateRateLimitPolicyFormat == "json" {
		outputBytes = jsonBytes
	} else {
		outputBytes, err = yaml.JSONToYAML(jsonBytes) // use `omitempty`'s from the json Marshal
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(outputBytes))
	return nil
}

func buildRateLimitPolicy(doc *openapi3.T) *kuadrantapiv1beta2.RateLimitPolicy {
	routeMeta := gatewayapi.HTTPRouteObjectMetaFromOAS(doc)

	rlp := &kuadrantapiv1beta2.RateLimitPolicy{
		TypeMeta: v1.TypeMeta{
			APIVersion: "kuadrant.io/v1beta2",
			Kind:       "RateLimitPolicy",
		},
		ObjectMeta: kuadrantapi.RateLimitPolicyObjectMetaFromOAS(doc),
		Spec: kuadrantapiv1beta2.RateLimitPolicySpec{
			TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
				Group: gatewayapiv1beta1.Group("gateway.networking.k8s.io"),
				Kind:  gatewayapiv1beta1.Kind("HTTPRoute"),
				Name:  gatewayapiv1beta1.ObjectName(routeMeta.Name),
			},
			Limits: kuadrantapi.RateLimitPolicyLimitsFromOAS(doc),
		},
	}

	if routeMeta.Namespace != "" {
		rlp.Spec.TargetRef.Namespace = &[]gatewayapiv1beta1.Namespace{
			gatewayapiv1beta1.Namespace(routeMeta.Namespace),
		}[0]
	}

	return rlp
}
