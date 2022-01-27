package gatewayapi

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func HTTPRouteMatchesFromOAS(doc *openapi3.T) ([]gatewayapiv1alpha2.HTTPRouteMatch, error) {
	httpRouteMatches := []gatewayapiv1alpha2.HTTPRouteMatch{}
	pathMatchPathPrefix := gatewayapiv1alpha2.PathMatchPathPrefix
	pathMatchExactPath := gatewayapiv1alpha2.PathMatchExact

	for path, pathItem := range doc.Paths {

		headers := []gatewayapiv1alpha2.HTTPHeaderMatch{}
		queryParams := []gatewayapiv1alpha2.HTTPQueryParamMatch{}
		headers, queryParams = addRuleMatcherFromParams(pathItem.Parameters, headers, queryParams)

		for verb, operation := range pathItem.Operations() {

			headers, queryParams = addRuleMatcherFromParams(operation.Parameters, headers, queryParams)

			pathMatch := &pathMatchExactPath
			for _, param := range queryParams {
				if param.Name == strings.ToLower(string(pathMatchPathPrefix)) {
					pathMatch = &pathMatchPathPrefix
				}
			}
			pathValue := path
			path := &gatewayapiv1alpha2.HTTPPathMatch{
				Type:  pathMatch,
				Value: &pathValue,
			}

			httpMethod := gatewayapiv1alpha2.HTTPMethod(verb)
			httpRouteMatches = append(httpRouteMatches, gatewayapiv1alpha2.HTTPRouteMatch{
				Method:      &httpMethod,
				Path:        path,
				Headers:     headers,
				QueryParams: queryParams,
			})
		}
	}

	return httpRouteMatches, nil
}

func addRuleMatcherFromParams(params openapi3.Parameters, headers []gatewayapiv1alpha2.HTTPHeaderMatch, queryParams []gatewayapiv1alpha2.HTTPQueryParamMatch) ([]gatewayapiv1alpha2.HTTPHeaderMatch, []gatewayapiv1alpha2.HTTPQueryParamMatch) {
	headerMatchType := gatewayapiv1alpha2.HeaderMatchExact
	queryParamMatchExact := gatewayapiv1alpha2.QueryParamMatchExact

	for _, parameter := range params {
		if !parameter.Value.Required {
			continue
		}

		if parameter.Value.In == openapi3.ParameterInHeader {
			headers = append(headers, gatewayapiv1alpha2.HTTPHeaderMatch{
				Type: &headerMatchType,
				Name: gatewayapiv1alpha2.HTTPHeaderName(parameter.Value.Name),
			})
		}
		if parameter.Value.In == openapi3.ParameterInQuery {
			queryParams = append(queryParams, gatewayapiv1alpha2.HTTPQueryParamMatch{
				Type: &queryParamMatchExact,
				Name: parameter.Value.Name,
			})
		}
	}

	return headers, queryParams
}
