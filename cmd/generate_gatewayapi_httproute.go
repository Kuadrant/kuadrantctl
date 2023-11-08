package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	generateGatewayAPIHTTPRouteOAS          string
	generateGatewayAPIHTTPRouteHost         string
	generateGatewayAPIHTTPRouteSvcName      string
	generateGatewayAPIHTTPRouteSvcNamespace string
	generateGatewayAPIHTTPRouteSvcPort      int32
	generateGatewayAPIHTTPRouteGateways     []string
)

//kuadrantctl generate istio virtualservice --namespace myns --oas petstore.yaml --public-host www.kuadrant.io --service-name myservice --gateway kuadrant-gateway
// --namespace myns
// --service-name myservice
// --public-host www.kuadrant.io
// --gateway kuadrant-gateway
// -- service-port 80

func generateGatewayApiHttpRouteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "httproute",
		Short: "Generate Gateway API HTTPRoute from OpenAPI 3.x",
		Long:  "Generate Gateway API HTTPRoute from OpenAPI 3.x",
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateGatewayApiHttpRoute(cmd, args)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateGatewayAPIHTTPRouteOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	// service ref
	cmd.Flags().StringVar(&generateGatewayAPIHTTPRouteSvcName, "service-name", "", "Service name (required)")
	err = cmd.MarkFlagRequired("service-name")
	if err != nil {
		panic(err)
	}

	// service namespace
	cmd.Flags().StringVarP(&generateGatewayAPIHTTPRouteSvcNamespace, "namespace", "n", "", "Service namespace (required)")
	err = cmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}

	// service host
	cmd.Flags().StringVar(&generateGatewayAPIHTTPRouteHost, "public-host", "", "Public host (required)")
	err = cmd.MarkFlagRequired("public-host")
	if err != nil {
		panic(err)
	}

	// service port
	cmd.Flags().Int32VarP(&generateGatewayAPIHTTPRouteSvcPort, "port", "p", 80, "Service Port (required)")

	// gateway
	cmd.Flags().StringSliceVar(&generateGatewayAPIHTTPRouteGateways, "gateway", []string{}, "Gateways (required)")
	err = cmd.MarkFlagRequired("gateway")
	if err != nil {
		panic(err)
	}

	return cmd
}

func generateGatewayApiHttpRoute(cmd *cobra.Command, args []string) error {
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

	httpRoute, err := generateGatewayAPIHTTPRoute(cmd, doc)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(httpRoute)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func generateGatewayAPIHTTPRoute(cmd *cobra.Command, doc *openapi3.T) (*gatewayapiv1alpha2.HTTPRoute, error) {

	//loop through gateway
	// https://github.com/getkin/kin-openapi
	gatewaysRef := []gatewayapiv1beta1.ParentReference{}
	for _, gateway := range generateGatewayAPIHTTPRouteGateways {
		gatewaysRef = append(gatewaysRef, gatewayapiv1beta1.ParentReference{
			Name: gatewayapiv1alpha2.ObjectName(gateway),
		})
	}

	port := gatewayapiv1alpha2.PortNumber(generateGatewayAPIHTTPRouteSvcPort)
	service := fmt.Sprintf("%s.%s.svc", generateGatewayAPIHTTPRouteSvcName, generateGatewayAPIHTTPRouteSvcNamespace)
	matches, err := gatewayapi.HTTPRouteMatchesFromOAS(doc)
	if err != nil {
		return nil, err
	}

	httpRoute := gatewayapiv1alpha2.HTTPRoute{
		TypeMeta: v1.TypeMeta{
			Kind:       "HTTPRoute",
			APIVersion: "gateway.networking.k8s.io/v1alpha2",
		},
		Spec: gatewayapiv1alpha2.HTTPRouteSpec{
			CommonRouteSpec: gatewayapiv1beta1.CommonRouteSpec{
				ParentRefs: gatewaysRef,
			},
			Hostnames: []gatewayapiv1alpha2.Hostname{
				gatewayapiv1alpha2.Hostname(generateGatewayAPIHTTPRouteHost),
			},
			Rules: []gatewayapiv1alpha2.HTTPRouteRule{
				{
					BackendRefs: []gatewayapiv1alpha2.HTTPBackendRef{
						{
							BackendRef: gatewayapiv1alpha2.BackendRef{
								BackendObjectReference: gatewayapiv1alpha2.BackendObjectReference{
									Name: gatewayapiv1alpha2.ObjectName(service),
									Port: &port,
								},
							},
						},
					},
					Matches: matches,
				},
			},
		},
	}
	return &httpRoute, nil
}
