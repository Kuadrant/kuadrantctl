package gatewayapi

import (
	"fmt"

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

	basePath, err := utils.BasePathFromOpenAPI(doc)
	if err != nil {
		panic(err)
	}

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
				// not enabled for the operation
				continue
			}

			// default backendrefs at the path level
			backendRefs := kuadrantPathExtension.BackendRefs
			if len(kuadrantOperationExtension.BackendRefs) > 0 {
				backendRefs = kuadrantOperationExtension.BackendRefs
			}

			rules = append(rules, buildHTTPRouteRule(basePath, path, pathItem, verb, operation, backendRefs))
		}
	}

	if len(rules) == 0 {
		return nil
	}

	return rules
}

func ExtractLabelsFromOAS(doc *openapi3.T) (map[string]string, bool) {
	if doc.Info == nil || doc.Info.Extensions == nil {
		return nil, false
	}

	if extension, ok := doc.Info.Extensions["x-kuadrant"]; ok {
		if extensionMap, ok := extension.(map[string]interface{}); ok {
			if route, ok := extensionMap["route"].(map[string]interface{}); ok {
				if labelsInterface, ok := route["labels"]; ok {
					labels := make(map[string]string)
					for key, value := range labelsInterface.(map[string]interface{}) {
						labels[key] = fmt.Sprint(value)
					}
					return labels, true
				}
			}
		}
	}

	return nil, false
}

func buildHTTPRouteRule(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, backendRefs []gatewayapiv1beta1.HTTPBackendRef) gatewayapiv1beta1.HTTPRouteRule {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op)

	return gatewayapiv1beta1.HTTPRouteRule{
		BackendRefs: backendRefs,
		Matches:     []gatewayapiv1beta1.HTTPRouteMatch{match},
	}
}
