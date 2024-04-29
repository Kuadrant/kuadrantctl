package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/kuadrantapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

//kuadrantctl generate kuadrant authpolicy --oas [OAS_FILE_PATH | OAS_URL | @]

func generateKuadrantAuthPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authpolicy",
		Short: "Generate Kuadrant AuthPolicy from OpenAPI 3.0.X",
		Long:  "Generate Kuadrant AuthPolicy from OpenAPI 3.0.X",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, err := cmd.Flags().GetString("output-format")
			if err != nil {
				return err
			}

			oasPath, err := cmd.Flags().GetString("oas")
			if err != nil {
				return err
			}

			return runGenerateKuadrantAuthPolicy(cmd, oasPath, outputFormat)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateGatewayAPIHTTPRouteOAS, "oas", "", "Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required)")
	cmd.Flags().StringP("output-format", "o", "yaml", "Output format: 'yaml' or 'json'. Default: yaml")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	return cmd
}

func runGenerateKuadrantAuthPolicy(cmd *cobra.Command, oasPath, outputFormat string) error {
	oasDataRaw, err := utils.ReadExternalResource(oasPath)
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

	ap := buildAuthPolicy(doc)

	var outputBytes []byte
	if outputFormat == "json" {
		outputBytes, err = json.Marshal(ap)
	} else { // default to YAML if not explicitly JSON
		outputBytes, err = yaml.Marshal(ap)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(outputBytes))
	return nil
}

func buildAuthPolicy(doc *openapi3.T) *kuadrantapiv1beta2.AuthPolicy {
	routeMeta := gatewayapi.HTTPRouteObjectMetaFromOAS(doc)

	ap := &kuadrantapiv1beta2.AuthPolicy{
		TypeMeta: v1.TypeMeta{
			APIVersion: "kuadrant.io/v1beta2",
			Kind:       "AuthPolicy",
		},
		ObjectMeta: kuadrantapi.AuthPolicyObjectMetaFromOAS(doc),
		Spec: kuadrantapiv1beta2.AuthPolicySpec{
			TargetRef: gatewayapiv1alpha2.PolicyTargetReference{
				Group: gatewayapiv1beta1.Group("gateway.networking.k8s.io"),
				Kind:  gatewayapiv1beta1.Kind("HTTPRoute"),
				Name:  gatewayapiv1beta1.ObjectName(routeMeta.Name),
			},
			// Currently only authentication rules enforced
			AuthScheme: kuadrantapiv1beta2.AuthSchemeSpec{
				Authentication: kuadrantapi.AuthPolicyAuthenticationSchemeFromOAS(doc),
			},
			RouteSelectors: kuadrantapi.AuthPolicyTopRouteSelectorsFromOAS(doc),
		},
	}

	if routeMeta.Namespace != "" {
		ap.Spec.TargetRef.Namespace = &[]gatewayapiv1beta1.Namespace{
			gatewayapiv1beta1.Namespace(routeMeta.Namespace),
		}[0]
	}

	return ap
}
