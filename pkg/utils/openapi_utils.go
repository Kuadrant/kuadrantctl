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
package utils

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	// NonWordCharRegexp not word characters (== [^0-9A-Za-z_])
	NonWordCharRegexp = regexp.MustCompile(`\W`)

	// TemplateRegexp used to render openapi server URLs
	TemplateRegexp = regexp.MustCompile(`{([\w]+)}`)

	// NonAlphanumRegexp not alphanumeric
	NonAlphanumRegexp = regexp.MustCompile(`[^0-9A-Za-z]`)
)

func MethodNameFromOpenAPIOperation(path, opVerb string, op *openapi3.Operation) string {
	sanitizedPath := NonWordCharRegexp.ReplaceAllString(path, "")

	name := fmt.Sprintf("%s%s", opVerb, sanitizedPath)

	if op.OperationID != "" {
		name = op.OperationID
	}
	return name
}

func MethodSystemNameFromOpenAPIOperation(path, opVerb string, op *openapi3.Operation) string {
	nameToLower := strings.ToLower(MethodNameFromOpenAPIOperation(path, opVerb, op))
	return NonWordCharRegexp.ReplaceAllString(nameToLower, "_")
}

func K8sNameFromOpenAPITitle(doc *openapi3.T) string {
	openapiTitleToLower := strings.ToLower(doc.Info.Title)
	return NonAlphanumRegexp.ReplaceAllString(openapiTitleToLower, "")
}

func RenderOpenAPIServerURLStr(server *openapi3.Server) (string, error) {
	if server == nil {
		return "", nil
	}

	data := &struct {
		Data map[string]string
	}{
		map[string]string{},
	}

	for variableName, variable := range server.Variables {
		data.Data[variableName] = variable.Default
	}

	urlTemplate := TemplateRegexp.ReplaceAllString(server.URL, `{{ index .Data "$1" }}`)

	tObj, err := template.New(server.URL).Parse(urlTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = tObj.Execute(&tpl, data)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func RenderOpenAPIServerURL(server *openapi3.Server) (*url.URL, error) {
	serverURLStr, err := RenderOpenAPIServerURLStr(server)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(serverURLStr)
	if err != nil {
		return nil, err
	}

	return serverURL, nil
}

func FirstServerFromOpenAPI(doc *openapi3.T) *openapi3.Server {
	if doc == nil {
		return nil
	}

	// take only first server
	// From https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md
	//   If the servers property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	server := &openapi3.Server{
		URL:       `/`,
		Variables: map[string]*openapi3.ServerVariable{},
	}

	if len(doc.Servers) > 0 {
		server = doc.Servers[0]
	}

	return server
}

func BasePathFromOpenAPI(doc *openapi3.T) (string, error) {
	server := FirstServerFromOpenAPI(doc)
	serverURL, err := RenderOpenAPIServerURL(server)
	if err != nil {
		return "", err
	}

	return serverURL.Path, nil
}
