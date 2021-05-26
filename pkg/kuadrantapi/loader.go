/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kuadrantapi

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

// Loader helps deserialize an OpenAPIv3 document
type Loader struct {
}

// NewLoader returns an empty Loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromResource loads a API spec from an external OpenAPI resource
// Currently implemented data streams:
// - '-' for STDIN
// - URLs (HTTP[S])
// - Files
func (loader *Loader) LoadFromResource(resource string) (*kctlrv1beta1.API, error) {
	data, err := utils.ReadExternalResource(resource)
	if err != nil {
		return nil, err
	}

	openapiLoader := openapi3.NewLoader()
	doc, err := openapiLoader.LoadFromData(data)
	if err != nil {
		return nil, err
	}

	// TODO(eastizle): optional flag for validation
	err = doc.Validate(openapiLoader.Context)
	if err != nil {
		return nil, err
	}

	return loader.LoadFromDoc(doc)
}

// LoadFromDoc loads a API spec from an OpenAPI doc
func (loader *Loader) LoadFromDoc(doc *openapi3.T) (*kctlrv1beta1.API, error) {
	api := &kctlrv1beta1.API{
		TypeMeta: metav1.TypeMeta{
			Kind:       "API",
			APIVersion: kctlrv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.K8sNameFromOpenAPITitle(doc),
		},
	}

	err := loader.loadServers(doc, api)
	if err != nil {
		return nil, err
	}
	err = loader.loadOperations(doc, api)
	if err != nil {
		return nil, err
	}
	err = loader.loadSecuritySchemes(doc, api)
	if err != nil {
		return nil, err
	}
	err = loader.loadGlobalSecurityRequirements(doc, api)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (loader *Loader) loadServers(doc *openapi3.T, api *kctlrv1beta1.API) error {
	for _, server := range doc.Servers {
		serverURL, err := utils.RenderOpenAPIServerURL(server)
		if err != nil {
			return err
		}
		api.Spec.Hosts = append(api.Spec.Hosts, serverURL.Host)
	}

	return nil
}

func (loader *Loader) loadOperations(doc *openapi3.T, api *kctlrv1beta1.API) error {
	basePath, err := utils.BasePathFromOpenAPI(doc)
	if err != nil {
		return err
	}

	for path, pathItem := range doc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			apiOperation := &kctlrv1beta1.Operation{
				Name:   utils.MethodSystemNameFromOpenAPIOperation(path, opVerb, operation),
				Path:   fmt.Sprintf("%s%s", basePath, path),
				Method: opVerb,
			}
			if operation.Security != nil {
				// TODO(eastizle) Current version of API (from github.com/kuadrant/kuadrant-controller#0.0.1-pre)
				// data type cannot accomodate sec requirements for operations
			}

			api.Spec.Operations = append(api.Spec.Operations, apiOperation)
		}
	}

	return nil
}

func (loader *Loader) loadSecuritySchemes(doc *openapi3.T, api *kctlrv1beta1.API) error {
	for secSchemeName, secScheme := range doc.Components.SecuritySchemes {
		kapiSecSchemeObj := &kctlrv1beta1.SecurityScheme{
			Name: secSchemeName,
		}

		switch secScheme.Value.Type {
		case "apiKey":
			kapiSecSchemeObj.APIKeyAuth = &kctlrv1beta1.APIKeyAuth{
				Location: secScheme.Value.In,
				Name:     secScheme.Value.Name,
			}
		case "openIdConnect":
			kapiSecSchemeObj.OpenIDConnectAuth = &kctlrv1beta1.OpenIDConnectAuth{
				URL: secScheme.Value.OpenIdConnectUrl,
			}
		default:
			return fmt.Errorf("Unexpected security scheme type found: %s. Supported values are: %s",
				secScheme.Value.Type, []string{"apiKey", "openIdConnect"})
		}
		api.Spec.SecurityScheme = append(api.Spec.SecurityScheme, kapiSecSchemeObj)
	}

	return nil
}

func (loader *Loader) loadGlobalSecurityRequirements(doc *openapi3.T, api *kctlrv1beta1.API) error {
	// TODO(eastizle) Current version of API (from github.com/kuadrant/kuadrant-controller#0.0.1-pre)
	// data type cannot accomodate global sec requirements

	return nil
}
