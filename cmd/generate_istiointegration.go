package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	istioapi "istio.io/api/networking/v1beta1"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	istioutils "github.com/kuadrant/kuadrantctl/pkg/istio"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	generateIstioOAS              string
	generateIstioPublicHost       string
	generateIstioServiceName      string
	generateIstioServiceNamespace string
	generateIstioServicePort      int32
)

func generateIstioIntegrationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istiointegration",
		Short: "Generate Istio IstioIntegration from OpenAPI 3.x",
		Long:  "Generate Istio IstioIntegration from OpenAPI 3.x",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateIstioIntegration(cmd, args)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateIstioOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	// public host
	cmd.Flags().StringVar(&generateIstioPublicHost, "public-host", "", "The address used by a client when attempting to connect to a service (required)")
	err = cmd.MarkFlagRequired("public-host")
	if err != nil {
		panic(err)
	}

	// service name
	cmd.Flags().StringVar(&generateIstioServiceName, "service-name", "", "Service name (required)")
	err = cmd.MarkFlagRequired("service-name")
	if err != nil {
		panic(err)
	}

	// service namespace
	cmd.Flags().StringVarP(&generateIstioServiceNamespace, "namespace", "n", "", "Service namespace (required)")
	err = cmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}

	// service port
	cmd.Flags().Int32VarP(&generateIstioServicePort, "service-port", "p", 80, "Service port (required)")
	err = cmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}

	return cmd
}

func runGenerateIstioIntegration(cmd *cobra.Command, args []string) error {
	dataRaw, err := utils.ReadExternalResource(generateIstioOAS)
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

	err = generateIstioVirtualService(cmd, doc)
	if err != nil {
		return err
	}

	return nil
}

func generateIstioVirtualService(cmd *cobra.Command, doc *openapi3.T) error {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return err
	}

	destination := &istioapi.Destination{
		Host: fmt.Sprintf("%s.%s.svc", generateIstioServiceName, generateIstioServiceNamespace),
		Port: &istioapi.PortSelector{Number: uint32(generateIstioServicePort)},
	}

	httpRoutes, err := istioutils.HTTPRoutesFromOpenAPI(doc, destination)
	if err != nil {
		return err
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
		Spec: istioapi.VirtualService{
			Hosts: []string{generateIstioPublicHost},
			Http:  httpRoutes,
		},
	}

	jsonData, err := json.Marshal(vs)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))

	return nil
}
