package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	istionetworkingapi "istio.io/api/networking/v1beta1"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	istioutils "github.com/kuadrant/kuadrantctl/pkg/istio"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	generateIstioVSOAS              string
	generateIstioVSPublicHost       string
	generateIstioVSServiceName      string
	generateIstioVSServiceNamespace string
	generateIstioVSServicePort      int32
	generateIstioVSGateways         []string
	generateIstioVSPrefixMatch      bool
)

func generateIstioVirtualServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "virtualservice",
		Short: "Generate Istio VirtualService from OpenAPI 3.x",
		Long:  "Generate Istio VirtualService from OpenAPI 3.x",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rungenerateistiovirtualservicecommand(cmd, args)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateIstioVSOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	// public host
	cmd.Flags().StringVar(&generateIstioVSPublicHost, "public-host", "", "The address used by a client when attempting to connect to a service (required)")
	err = cmd.MarkFlagRequired("public-host")
	if err != nil {
		panic(err)
	}

	// service name
	cmd.Flags().StringVar(&generateIstioVSServiceName, "service-name", "", "Service name (required)")
	err = cmd.MarkFlagRequired("service-name")
	if err != nil {
		panic(err)
	}

	// service namespace
	cmd.Flags().StringVarP(&generateIstioVSServiceNamespace, "service-namespace", "", "", "Service namespace (required)")
	err = cmd.MarkFlagRequired("service-namespace")
	if err != nil {
		panic(err)
	}

	// service port
	cmd.Flags().Int32VarP(&generateIstioVSServicePort, "service-port", "p", 80, "Service port")

	// gateways
	cmd.Flags().StringSliceVar(&generateIstioVSGateways, "gateway", []string{}, "Gateways (required)")
	err = cmd.MarkFlagRequired("gateway")
	if err != nil {
		panic(err)
	}

	// exact match
	cmd.Flags().BoolVar(&generateIstioVSPrefixMatch, "path-prefix-match", false, "Path match type (defaults to exact match type)")

	return cmd
}

func rungenerateistiovirtualservicecommand(cmd *cobra.Command, args []string) error {
	dataRaw, err := utils.ReadExternalResource(generateIstioVSOAS)
	if err != nil {
		return err
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData(dataRaw)
	if err != nil {
		return err
	}

	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		return fmt.Errorf("OpenAPI validation error: %w", err)
	}

	vs, err := generateIstioVirtualService(cmd, doc)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(vs)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func generateIstioVirtualService(cmd *cobra.Command, doc *openapi3.T) (*istionetworking.VirtualService, error) {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	destination := &istionetworkingapi.Destination{
		Host: fmt.Sprintf("%s.%s.svc", generateIstioVSServiceName, generateIstioVSServiceNamespace),
		Port: &istionetworkingapi.PortSelector{Number: uint32(generateIstioVSServicePort)},
	}

	httpRoutes, err := istioutils.HTTPRoutesFromOpenAPI(doc, destination, generateIstioVSPrefixMatch)
	if err != nil {
		return nil, err
	}

	vs := &istionetworking.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			// Missing namespace
			Name: objectName,
		},
		Spec: istionetworkingapi.VirtualService{
			Gateways: generateIstioVSGateways,
			Hosts:    []string{generateIstioVSPublicHost},
			Http:     httpRoutes,
		},
	}

	return vs, nil
}
