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

// LoadFromDoc loads a spec from an OpenAPI doc
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

	_, err := utils.BasePathFromOpenAPI(doc)
	if err != nil {
		return nil, err
	}

	for _, server := range doc.Servers {
		serverURL, err := utils.RenderOpenAPIServerURL(server)
		if err != nil {
			return nil, err
		}
		api.Spec.Hosts = append(api.Spec.Hosts, serverURL.Host)
	}

	for path, pathItem := range doc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			apiOperation := &kctlrv1beta1.Operation{
				Name:   utils.MethodSystemNameFromOpenAPIOperation(path, opVerb, operation),
				Path:   path,
				Method: opVerb,
			}
			api.Spec.Operations = append(api.Spec.Operations, apiOperation)
		}
	}

	return api, nil
}
