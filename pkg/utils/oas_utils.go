package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	// NonWordCharRegexp not word characters (== [^0-9A-Za-z_])
	NonWordCharRegexp = regexp.MustCompile(`\W`)
	// TemplateRegexp used to render openapi server URLs
	TemplateRegexp = regexp.MustCompile(`{([\w]+)}`)
	// LastSlashRegexp matches the last slash
	LastSlashRegexp = regexp.MustCompile(`/$`)
)

func FirstServerFromOpenAPI(obj *openapi3.T) *openapi3.Server {
	if obj == nil {
		return nil
	}

	// take only first server
	// From https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md
	//   If the servers property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	server := &openapi3.Server{
		URL:       `/`,
		Variables: map[string]*openapi3.ServerVariable{},
	}

	// Current constraint: only read the first item when there are multiple servers
	// Maybe this should be user provided setting
	if len(obj.Servers) > 0 {
		server = obj.Servers[0]
	}

	return server
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

func BasePathFromOpenAPI(obj *openapi3.T) (string, error) {
	server := FirstServerFromOpenAPI(obj)
	serverURL, err := RenderOpenAPIServerURL(server)
	if err != nil {
		return "", err
	}

	return serverURL.Path, nil
}

func OpenAPIMatcherFromOASOperations(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1beta1.PathMatchType) gatewayapiv1beta1.HTTPRouteMatch {
	// remove the last slash of the Base Path
	sanitizedBasePath := LastSlashRegexp.ReplaceAllString(basePath, "")

	//  According OAS 3.0: path MUST begin with a slash
	matchPath := fmt.Sprintf("%s%s", sanitizedBasePath, path)

	pathHeadersMatch := headersMatchFromParams(pathItem.Parameters)
	operationHeadersMatch := headersMatchFromParams(op.Parameters)

	// default headersMatch at the path level
	headersMatch := pathHeadersMatch
	if len(operationHeadersMatch) > 0 {
		headersMatch = operationHeadersMatch
	}

	pathQueryParamsMatch := queryParamsMatchFromParams(pathItem.Parameters)
	operationQueryParamsMatch := queryParamsMatchFromParams(op.Parameters)

	// default queryParams at the path level
	queryParams := pathQueryParamsMatch
	if len(operationQueryParamsMatch) > 0 {
		queryParams = operationQueryParamsMatch
	}

	return gatewayapiv1beta1.HTTPRouteMatch{
		Method: &[]gatewayapiv1beta1.HTTPMethod{gatewayapiv1beta1.HTTPMethod(verb)}[0],
		Path: &gatewayapiv1beta1.HTTPPathMatch{
			Type:  &pathMatchType,
			Value: &[]string{matchPath}[0],
		},
		Headers:     headersMatch,
		QueryParams: queryParams,
	}
}

func headersMatchFromParams(params openapi3.Parameters) []gatewayapiv1beta1.HTTPHeaderMatch {
	matches := make([]gatewayapiv1beta1.HTTPHeaderMatch, 0)

	for _, parameter := range params {
		if !parameter.Value.Required {
			continue
		}

		if parameter.Value.In == openapi3.ParameterInHeader {
			matches = append(matches, gatewayapiv1beta1.HTTPHeaderMatch{
				Type: &[]gatewayapiv1beta1.HeaderMatchType{gatewayapiv1beta1.HeaderMatchExact}[0],
				Name: gatewayapiv1beta1.HTTPHeaderName(parameter.Value.Name),
			})
		}
	}

	if len(matches) == 0 {
		return nil
	}

	return matches
}

func queryParamsMatchFromParams(params openapi3.Parameters) []gatewayapiv1beta1.HTTPQueryParamMatch {
	matches := make([]gatewayapiv1beta1.HTTPQueryParamMatch, 0)

	for _, parameter := range params {
		if !parameter.Value.Required {
			continue
		}

		if parameter.Value.In == openapi3.ParameterInQuery {
			matches = append(matches, gatewayapiv1beta1.HTTPQueryParamMatch{
				Type: &[]gatewayapiv1beta1.QueryParamMatchType{gatewayapiv1beta1.QueryParamMatchExact}[0],
				Name: parameter.Value.Name,
			})
		}
	}

	if len(matches) == 0 {
		return nil
	}

	return matches

}

func OpenAPIOperationName(path, opVerb string, op *openapi3.Operation) string {
	sanitizedPath := NonWordCharRegexp.ReplaceAllString(path, "")

	name := fmt.Sprintf("%s%s", opVerb, sanitizedPath)

	if op.OperationID != "" {
		name = op.OperationID
	}

	return name
}
