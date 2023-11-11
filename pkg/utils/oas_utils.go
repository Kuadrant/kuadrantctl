package utils

import (
	"fmt"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	// NonWordCharRegexp not word characters (== [^0-9A-Za-z_])
	NonWordCharRegexp = regexp.MustCompile(`\W`)
)

func OpenAPIMatcherFromOASOperations(path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation) gatewayapiv1beta1.HTTPRouteMatch {
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
			// TODO(eguzki): consider other path match types like PathPrefix
			Type:  &[]gatewayapiv1beta1.PathMatchType{gatewayapiv1beta1.PathMatchExact}[0],
			Value: &[]string{path}[0],
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
