package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	generateGatewayAPIHTTPRouteOAS string
)

//kuadrantctl generate gatewayapi httproute --oas [OAS_FILE_PATH | OAS_URL | @]

func generateGatewayApiHttpRouteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "httproute",
		Short: "Generate Gateway API HTTPRoute from OpenAPI 3.0.X",
		Long:  "Generate Gateway API HTTPRoute from OpenAPI 3.0.X",
		RunE:  runGenerateGatewayApiHttpRoute,
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateGatewayAPIHTTPRouteOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR @ (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	return cmd
}

func runGenerateGatewayApiHttpRoute(cmd *cobra.Command, args []string) error {
	oasDataRaw, err := utils.ReadExternalResource(generateGatewayAPIHTTPRouteOAS)
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

	httpRoute := buildHTTPRoute(doc)

	jsonData, err := json.Marshal(httpRoute)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func buildHTTPRoute(doc *openapi3.T) *gatewayapiv1beta1.HTTPRoute {
	return &gatewayapiv1beta1.HTTPRoute{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1beta1",
			Kind:       "HTTPRoute",
		},
		ObjectMeta: gatewayapi.HTTPRouteObjectMetaFromOAS(doc),
		Spec: gatewayapiv1beta1.HTTPRouteSpec{
			CommonRouteSpec: gatewayapiv1beta1.CommonRouteSpec{
				ParentRefs: gatewayapi.HTTPRouteGatewayParentRefsFromOAS(doc),
			},
			Hostnames: gatewayapi.HTTPRouteHostnamesFromOAS(doc),
			Rules:     gatewayapi.HTTPRouteRulesFromOAS(doc),
		},
	}
}
