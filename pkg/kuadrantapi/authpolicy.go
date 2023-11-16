package kuadrantapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	authorinoapi "github.com/kuadrant/authorino/api/v1beta2"
	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func AuthPolicyObjectMetaFromOAS(doc *openapi3.T) metav1.ObjectMeta {
	return gatewayapi.HTTPRouteObjectMetaFromOAS(doc)
}

func buildAuthPolicyRouteSelectors(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation) []kuadrantapiv1beta2.RouteSelector {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op)

	return []kuadrantapiv1beta2.RouteSelector{
		{
			Matches: []gatewayapiv1beta1.HTTPRouteMatch{match},
		},
	}
}

func AuthPolicyAuthenticationSchemeFromOAS(doc *openapi3.T) map[string]kuadrantapiv1beta2.AuthenticationSpec {
	authentication := make(map[string]kuadrantapiv1beta2.AuthenticationSpec)

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
				//fmt.Printf("OUT not enabled: path: %s, method: %s\n", path, verb)
				continue
			}

			// Get operation level security requirements or fallback to global security requirements
			secRequirements := ptr.Deref(operation.Security, doc.Security)

			if len(secRequirements) == 0 {
				// no security
				continue
			}

			oidcScheme := findOIDCSecuritySchemesFromRequirements(doc, secRequirements)

			authName := utils.OpenAPIOperationName(path, verb, operation)

			authentication[authName] = kuadrantapiv1beta2.AuthenticationSpec{
				CommonAuthRuleSpec: kuadrantapiv1beta2.CommonAuthRuleSpec{
					RouteSelectors: buildAuthPolicyRouteSelectors(basePath, path, pathItem, verb, operation),
				},
				AuthenticationSpec: authorinoapi.AuthenticationSpec{
					AuthenticationMethodSpec: authorinoapi.AuthenticationMethodSpec{
						Jwt: &authorinoapi.JwtAuthenticationSpec{
							IssuerUrl: oidcScheme.OpenIdConnectUrl,
						},
					},
				},
			}
		}
	}

	if len(authentication) == 0 {
		return nil
	}

	return authentication
}

func findOIDCSecuritySchemesFromRequirements(doc *openapi3.T, secRequirements openapi3.SecurityRequirements) *openapi3.SecurityScheme {
	for _, secReq := range secRequirements {
		for secReqItemName := range secReq {
			secScheme, ok := doc.Components.SecuritySchemes[secReqItemName]
			if !ok {
				// should never happen. OpenAPI validation should detect this issue
				continue
			}
			if secScheme == nil || secScheme.Value == nil {
				continue
			}
			// Ref https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#fixed-fields-23
			if secScheme.Value.Type == "openIdConnect" {
				return secScheme.Value
			}
		}
	}

	return nil
}
