package gatewayapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func HTTPRouteMatchesFromOAS(doc *openapi3.T) ([]gatewayapiv1alpha2.HTTPRouteMatch, error) {
	httpRouteMatches := []gatewayapiv1alpha2.HTTPRouteMatch{}
	pathMatchExactPath := gatewayapiv1beta1.PathMatchExact

	for path, pathItem := range doc.Paths {

		headers := []gatewayapiv1alpha2.HTTPHeaderMatch{}
		queryParams := []gatewayapiv1alpha2.HTTPQueryParamMatch{}
		headers, queryParams = addRuleMatcherFromParams(pathItem.Parameters, headers, queryParams)

		for verb, operation := range pathItem.Operations() {

			headers, queryParams = addRuleMatcherFromParams(operation.Parameters, headers, queryParams)

			pathValue := path
			httpMethod := gatewayapiv1alpha2.HTTPMethod(verb)
			httpRouteMatches = append(httpRouteMatches, gatewayapiv1alpha2.HTTPRouteMatch{
				Method: &httpMethod,
				Path: &gatewayapiv1alpha2.HTTPPathMatch{
					Type:  &pathMatchExactPath,
					Value: &pathValue,
				},
				Headers:     headers,
				QueryParams: queryParams,
			})
		}
	}

	return httpRouteMatches, nil
}

func addRuleMatcherFromParams(params openapi3.Parameters, headers []gatewayapiv1alpha2.HTTPHeaderMatch, queryParams []gatewayapiv1alpha2.HTTPQueryParamMatch) ([]gatewayapiv1alpha2.HTTPHeaderMatch, []gatewayapiv1alpha2.HTTPQueryParamMatch) {
	headerMatchType := gatewayapiv1beta1.HeaderMatchExact
	queryParamMatchExact := gatewayapiv1beta1.QueryParamMatchExact

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
