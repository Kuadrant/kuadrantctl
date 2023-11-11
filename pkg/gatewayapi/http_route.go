package gatewayapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func HTTPRouteObjectMetaFromOAS(doc *openapi3.T) metav1.ObjectMeta {
	if doc.Info == nil {
		return metav1.ObjectMeta{}
	}

	kuadrantInfoExtension, err := utils.NewKuadrantOASInfoExtension(doc.Info)
	if err != nil {
		panic(err)
	}

	if kuadrantInfoExtension.Route == nil {
		panic("info kuadrant extension route not found")
	}

	if kuadrantInfoExtension.Route.Name == nil {
		panic("info kuadrant extension route name not found")
	}

	om := metav1.ObjectMeta{Name: *kuadrantInfoExtension.Route.Name}

	if kuadrantInfoExtension.Route.Namespace != nil {
		om.Namespace = *kuadrantInfoExtension.Route.Namespace
	}

	return om
}

func HTTPRouteGatewayParentRefsFromOAS(doc *openapi3.T) []gatewayapiv1beta1.ParentReference {
	if doc.Info == nil {
		return nil
	}

	kuadrantInfoExtension, err := utils.NewKuadrantOASInfoExtension(doc.Info)
	if err != nil {
		panic(err)
	}

	if kuadrantInfoExtension.Route == nil {
		panic("info kuadrant extension route not found")
	}

	return kuadrantInfoExtension.Route.ParentRefs
}

func HTTPRouteHostnamesFromOAS(doc *openapi3.T) []gatewayapiv1beta1.Hostname {
	if doc.Info == nil {
		return nil
	}

	kuadrantInfoExtension, err := utils.NewKuadrantOASInfoExtension(doc.Info)
	if err != nil {
		panic(err)
	}

	if kuadrantInfoExtension.Route == nil {
		panic("info kuadrant extension route not found")
	}

	return kuadrantInfoExtension.Route.Hostnames
}

func HTTPRouteRulesFromOAS(doc *openapi3.T) []gatewayapiv1beta1.HTTPRouteRule {
	// Current implementation, one rule per operation
	// TODO(eguzki): consider about grouping operations as HTTPRouteMatch objects in fewer HTTPRouteRule objects
	rules := make([]gatewayapiv1beta1.HTTPRouteRule, 0)

	// Paths
	for path, pathItem := range doc.Paths {
		kuadrantPathExtension, err := utils.NewKuadrantOASPathExtension(pathItem)
		if err != nil {
			panic(err)
		}

		pathEnabled := kuadrantPathExtension.IsEnabled()

		// Operations
		for verb, operation := range pathItem.Operations() {
			kuadrantOperationExtension, err := utils.NewKuadrantOASOperationExtension(operation)
			if err != nil {
				panic(err)
			}

			if !ptr.Deref(kuadrantOperationExtension.Enable, pathEnabled) {
				// not enabled for the HTTPRoute
				continue
			}

			// default backendrefs at the path level
			backendRefs := kuadrantPathExtension.BackendRefs
			if len(kuadrantOperationExtension.BackendRefs) > 0 {
				backendRefs = kuadrantOperationExtension.BackendRefs
			}

			rules = append(rules, buildHTTPRouteRule(path, pathItem, verb, operation, backendRefs))
		}
	}

	if len(rules) == 0 {
		return nil
	}

	return rules
}

func buildHTTPRouteRule(path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, backendRefs []gatewayapiv1beta1.HTTPBackendRef) gatewayapiv1beta1.HTTPRouteRule {
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

	match := gatewayapiv1beta1.HTTPRouteMatch{
		Method: &[]gatewayapiv1beta1.HTTPMethod{gatewayapiv1beta1.HTTPMethod(verb)}[0],
		Path: &gatewayapiv1beta1.HTTPPathMatch{
			// TODO(eguzki): consider other path match types like PathPrefix
			Type:  &[]gatewayapiv1beta1.PathMatchType{gatewayapiv1beta1.PathMatchExact}[0],
			Value: &[]string{path}[0],
		},
		Headers:     headersMatch,
		QueryParams: queryParams,
	}

	return gatewayapiv1beta1.HTTPRouteRule{
		BackendRefs: backendRefs,
		Matches:     []gatewayapiv1beta1.HTTPRouteMatch{match},
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
