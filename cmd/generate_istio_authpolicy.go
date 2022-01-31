package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	istiosecurityapi "istio.io/api/security/v1beta1"
	istiotypeapi "istio.io/api/type/v1beta1"
	istiosecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuadrant/kuadrant-controller/pkg/common"
	istioutils "github.com/kuadrant/kuadrantctl/pkg/istio"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	generateIstioAPOAS         string
	generateIstioAPPublicHost  string
	generateIstioGatewayLabels []string
)

func generateIstioAuthorizationPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorizationpolicy",
		Short: "Generate Istio AuthorizationPolicy",
		Long:  "Generate Istio AuthorizationPolicy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateIstioAuthorizationPolicyCommand(cmd, args)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateIstioAPOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	// public host
	cmd.Flags().StringVar(&generateIstioAPPublicHost, "public-host", "", "The address used by a client when attempting to connect to a service (required)")
	err = cmd.MarkFlagRequired("public-host")
	if err != nil {
		panic(err)
	}

	// gateway labels
	cmd.Flags().StringSliceVar(&generateIstioGatewayLabels, "gateway-label", []string{}, "Gateway label (required)")
	err = cmd.MarkFlagRequired("gateway-label")
	if err != nil {
		panic(err)
	}

	return cmd
}

func runGenerateIstioAuthorizationPolicyCommand(cmd *cobra.Command, args []string) error {
	dataRaw, err := utils.ReadExternalResource(generateIstioAPOAS)
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

	ap, err := generateIstioAuthorizationPolicy(cmd, doc)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(ap)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))

	return nil
}

func generateIstioAuthorizationPolicy(cmd *cobra.Command, doc *openapi3.T) (*istiosecurity.AuthorizationPolicy, error) {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	matchLabels := map[string]string{}
	for idx := range generateIstioGatewayLabels {
		labels := strings.Split(generateIstioGatewayLabels[idx], "=")
		if len(labels) != 2 {
			return nil, fmt.Errorf("gateway labels have wrong syntax: %s", generateIstioGatewayLabels[idx])
		}

		matchLabels[labels[0]] = labels[1]
	}

	rules := istioutils.AuthorizationPolicyRulesFromOpenAPI(doc, generateIstioAPPublicHost)

	authPolicy := &istiosecurity.AuthorizationPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthorizationPolicy",
			APIVersion: "security.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			// Missing namespace
			Name: objectName,
		},
		Spec: istiosecurityapi.AuthorizationPolicy{
			Selector: &istiotypeapi.WorkloadSelector{
				MatchLabels: matchLabels,
			},
			Rules:  rules,
			Action: istiosecurityapi.AuthorizationPolicy_CUSTOM,
			ActionDetail: &istiosecurityapi.AuthorizationPolicy_Provider{
				Provider: &istiosecurityapi.AuthorizationPolicy_ExtensionProvider{
					Name: common.KuadrantAuthorizationProvider,
				},
			},
		},
	}

	return authPolicy, nil
}
