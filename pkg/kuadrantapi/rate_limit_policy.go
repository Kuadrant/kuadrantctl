package kuadrantapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	kuadrantapiv1 "github.com/kuadrant/kuadrant-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func RateLimitPolicyObjectMetaFromOAS(doc *openapi3.T) metav1.ObjectMeta {
	return gatewayapi.HTTPRouteObjectMetaFromOAS(doc)
}

func RateLimitPolicyLimitsFromOAS(doc *openapi3.T) map[string]kuadrantapiv1.Limit {
	// Current implementation, one limit per operation
	// TODO(eguzki): consider about grouping operations in fewer RLP limits

	limits := make(map[string]kuadrantapiv1.Limit)

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
				//fmt.Printf("OUT not enabled: path: %s, method: %s\n", path, verb)
				continue
			}

			// default backendrefs at the path level
			rateLimit := kuadrantPathExtension.RateLimit
			if kuadrantOperationExtension.RateLimit != nil {
				rateLimit = kuadrantOperationExtension.RateLimit
			}

			if rateLimit == nil {
				// no rate limit defined for this operation
				//fmt.Printf("OUT no rate limit defined: path: %s, method: %s\n", path, verb)
				continue
			}

			// default pathMatchType at the path level
			pathMatchType := ptr.Deref(
				kuadrantOperationExtension.PathMatchType,
				kuadrantPathExtension.GetPathMatchType(),
			)

			limitName := utils.OpenAPIOperationName(path, verb, operation)

			limits[limitName] = kuadrantapiv1.Limit{
				RouteSelectors: buildLimitRouteSelectors(basePath, path, pathItem, verb, operation, pathMatchType),
				When:           rateLimit.When,
				Counters:       rateLimit.Counters,
				Rates:          rateLimit.Rates,
			}
		}
	}

	if len(limits) == 0 {
		return nil
	}

	return limits
}

func buildLimitRouteSelectors(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1.PathMatchType) []kuadrantapiv1.RouteSelector {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op, pathMatchType)

	return []kuadrantapiv1.RouteSelector{
		{
			Matches: []gatewayapiv1.HTTPRouteMatch{match},
		},
	}
}
