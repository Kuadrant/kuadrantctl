package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	authorinov1beta1 "github.com/kuadrant/authorino/api/v1beta1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	authorinoutils "github.com/kuadrant/kuadrantctl/pkg/authorino"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	generateKuadrantAuthConfigOAS        string
	generateKuadrantAuthConfigPublicHost string
)

func generateKuadrantAuthconfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authconfig",
		Short: "Generate kuadrant authconfig from OpenAPI 3.x",
		Long:  "Generate kuadrant authconfig from OpenAPI 3.x",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateKuadrantAuthconfigCommand(cmd, args)
		},
	}

	// OpenAPI ref
	cmd.Flags().StringVar(&generateKuadrantAuthConfigOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)")
	err := cmd.MarkFlagRequired("oas")
	if err != nil {
		panic(err)
	}

	// public host
	cmd.Flags().StringVar(&generateKuadrantAuthConfigPublicHost, "public-host", "", "The address used by a client when attempting to connect to a service (required)")
	err = cmd.MarkFlagRequired("public-host")
	if err != nil {
		panic(err)
	}

	return cmd
}

func runGenerateKuadrantAuthconfigCommand(cmd *cobra.Command, args []string) error {
	dataRaw, err := utils.ReadExternalResource(generateKuadrantAuthConfigOAS)
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

	authConfig, err := generateKuadrantAuthConfig(cmd, doc)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}

func generateKuadrantAuthConfig(cmd *cobra.Command, doc *openapi3.T) (*authorinov1beta1.AuthConfig, error) {
	objectName, err := utils.K8sNameFromOpenAPITitle(doc)
	if err != nil {
		return nil, err
	}

	identityList, err := authorinoutils.AuthConfigIdentitiesFromOpenAPI(doc)
	if err != nil {
		return nil, err
	}

	metadataList, err := generateKuadrantAuthConfigMetadata(doc)
	if err != nil {
		return nil, err
	}

	authorizationList, err := generateKuadrantAuthConfigAuthorization(doc)
	if err != nil {
		return nil, err
	}

	responseList, err := generateKuadrantAuthConfigResponse(doc)
	if err != nil {
		return nil, err
	}

	authConfig := &authorinov1beta1.AuthConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AuthConfig",
			APIVersion: "authorino.kuadrant.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: objectName,
		},
		Spec: authorinov1beta1.AuthConfigSpec{
			Hosts:         []string{generateKuadrantAuthConfigPublicHost},
			Identity:      identityList,
			Metadata:      metadataList,
			Authorization: authorizationList,
			Response:      responseList,
			Patterns:      nil,
			Conditions:    nil,
		},
	}
	return authConfig, nil
}

func generateKuadrantAuthConfigMetadata(doc *openapi3.T) ([]*authorinov1beta1.Metadata, error) {
	return nil, nil
}

func generateKuadrantAuthConfigAuthorization(doc *openapi3.T) ([]*authorinov1beta1.Authorization, error) {
	return nil, nil
}

func generateKuadrantAuthConfigResponse(doc *openapi3.T) ([]*authorinov1beta1.Response, error) {
	return nil, nil
}
