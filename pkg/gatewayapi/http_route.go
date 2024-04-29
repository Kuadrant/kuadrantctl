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

	if kuadrantInfoExtension == nil {
		return metav1.ObjectMeta{}
	}

	if kuadrantInfoExtension.Route == nil {
		panic("info kuadrant extension route not found")
	}

	if kuadrantInfoExtension.Route.Name == nil {
		panic("info kuadrant extension route name not found")
	}

	om := metav1.ObjectMeta{
		Name:   *kuadrantInfoExtension.Route.Name,
		Labels: kuadrantInfoExtension.Route.Labels,
	}

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

	if kuadrantInfoExtension == nil {
		return nil
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

	if kuadrantInfoExtension == nil {
		return nil
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

func buildHTTPRouteRule(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, backendRefs []gatewayapiv1beta1.HTTPBackendRef, pathMatchType gatewayapiv1beta1.PathMatchType) gatewayapiv1beta1.HTTPRouteRule {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op, pathMatchType)

	return gatewayapiv1beta1.HTTPRouteRule{
		BackendRefs: backendRefs,
		Matches:     []gatewayapiv1beta1.HTTPRouteMatch{match},
	}
}
