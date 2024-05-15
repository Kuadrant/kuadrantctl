package gatewayapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func HTTPRouteObjectMetaFromOAS(doc *openapi3.T) metav1.ObjectMeta {
	kuadrantRootExtension, err := utils.NewKuadrantOASRootExtension(doc)
	if err != nil {
		panic(err)
	}

	if kuadrantRootExtension == nil {
		return metav1.ObjectMeta{}
	}

	if kuadrantRootExtension.Route == nil {
		panic("openapi root kuadrant extension route not found")
	}

	if kuadrantRootExtension.Route.Name == nil {
		panic("openapi root kuadrant extension route name not found")
	}

	om := metav1.ObjectMeta{
		Name:   *kuadrantRootExtension.Route.Name,
		Labels: kuadrantRootExtension.Route.Labels,
	}

	if kuadrantRootExtension.Route.Namespace != nil {
		om.Namespace = *kuadrantRootExtension.Route.Namespace
	}

	return om
}

func HTTPRouteGatewayParentRefsFromOAS(doc *openapi3.T) []gatewayapiv1.ParentReference {
	kuadrantRootExtension, err := utils.NewKuadrantOASRootExtension(doc)
	if err != nil {
		panic(err)
	}

	if kuadrantRootExtension == nil {
		return nil
	}

	if kuadrantRootExtension.Route == nil {
		panic("openapi root kuadrant extension route not found")
	}

	return kuadrantRootExtension.Route.ParentRefs
}

func HTTPRouteHostnamesFromOAS(doc *openapi3.T) []gatewayapiv1.Hostname {
	kuadrantRootExtension, err := utils.NewKuadrantOASRootExtension(doc)
	if err != nil {
		panic(err)
	}

	if kuadrantRootExtension == nil {
		return nil
	}

	if kuadrantRootExtension.Route == nil {
		panic("openapi root kuadrant extension route not found")
	}

	return kuadrantRootExtension.Route.Hostnames
}

func HTTPRouteRulesFromOAS(doc *openapi3.T) []gatewayapiv1.HTTPRouteRule {
	// Current implementation, one rule per operation
	// TODO(eguzki): consider about grouping operations as HTTPRouteMatch objects in fewer HTTPRouteRule objects
	rules := make([]gatewayapiv1.HTTPRouteRule, 0)

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

		// Operations
		for verb, operation := range pathItem.Operations() {
			kuadrantOperationExtension, err := utils.NewKuadrantOASOperationExtension(operation)
			if err != nil {
				panic(err)
			}

			if ptr.Deref(kuadrantOperationExtension.Disable, kuadrantPathExtension.IsDisabled()) {
				// not enabled for the operation
				continue
			}

			// default backendrefs at the path level
			backendRefs := kuadrantPathExtension.BackendRefs
			if len(kuadrantOperationExtension.BackendRefs) > 0 {
				backendRefs = kuadrantOperationExtension.BackendRefs
			}

			// default pathMatchType at the path level
			pathMatchType := ptr.Deref(
				kuadrantOperationExtension.PathMatchType,
				kuadrantPathExtension.GetPathMatchType(),
			)

			rules = append(rules, buildHTTPRouteRule(basePath, path, pathItem, verb, operation, backendRefs, pathMatchType))
		}
	}

	if len(rules) == 0 {
		return nil
	}

	return rules
}

func buildHTTPRouteRule(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, backendRefs []gatewayapiv1.HTTPBackendRef, pathMatchType gatewayapiv1.PathMatchType) gatewayapiv1.HTTPRouteRule {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op, pathMatchType)

	return gatewayapiv1.HTTPRouteRule{
		BackendRefs: backendRefs,
		Matches:     []gatewayapiv1.HTTPRouteMatch{match},
	}
}
