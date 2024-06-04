package kuadrantapi

import (
	"errors"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	authorinoapi "github.com/kuadrant/authorino/api/v1beta2"
	kuadrantapiv1beta2 "github.com/kuadrant/kuadrant-operator/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kuadrant/kuadrantctl/pkg/gatewayapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

const (
	APIKeySecretLabel = "kuadrant.io/apikeys-by"
)

func AuthPolicyObjectMetaFromOAS(doc *openapi3.T) metav1.ObjectMeta {
	return gatewayapi.HTTPRouteObjectMetaFromOAS(doc)
}

func buildAuthPolicyRouteSelectors(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1.PathMatchType) []kuadrantapiv1beta2.RouteSelector {
	match := utils.OpenAPIMatcherFromOASOperations(basePath, path, pathItem, verb, op, pathMatchType)

	return []kuadrantapiv1beta2.RouteSelector{
		{
			Matches: []gatewayapiv1.HTTPRouteMatch{match},
		},
	}
}

func AuthPolicyTopRouteSelectorsFromOAS(doc *openapi3.T) []kuadrantapiv1beta2.RouteSelector {
	routeSelectors := make([]kuadrantapiv1beta2.RouteSelector, 0)

	basePath, err := utils.BasePathFromOpenAPI(doc)
	if err != nil {
		panic(err)
	}

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

			// Get operation level security requirements or fallback to global security requirements
			secRequirements := ptr.Deref(operation.Security, doc.Security)

			// Top RouteSelectors define the matching rules to call external auth service
			// group together any routes that has at least one security requirement
			if len(secRequirements) == 0 {
				// no security
				continue
			}

			// default pathMatchType at the path level
			pathMatchType := ptr.Deref(
				kuadrantOperationExtension.PathMatchType,
				kuadrantPathExtension.GetPathMatchType(),
			)

			routeSelectors = append(routeSelectors, buildAuthPolicyRouteSelectors(basePath, path, pathItem, verb, operation, pathMatchType)...)
		}
	}

	if len(routeSelectors) == 0 {
		return nil
	}

	return routeSelectors
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

			// Get operation level security requirements or fallback to global security requirements
			secRequirements := ptr.Deref(operation.Security, doc.Security)

			if len(secRequirements) == 0 {
				// no security
				continue
			}

			// default pathMatchType at the path level
			pathMatchType := ptr.Deref(
				kuadrantOperationExtension.PathMatchType,
				kuadrantPathExtension.GetPathMatchType(),
			)

			operationAuthentication := buildOperationAuthentication(doc, basePath, path, pathItem, verb, operation, pathMatchType, secRequirements)

			// Aggregate auth methods per operation
			authentication = utils.MergeMaps(authentication, operationAuthentication)
		}
	}

	if len(authentication) == 0 {
		return nil
	}

	return authentication
}

func buildOperationAuthentication(doc *openapi3.T, basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1.PathMatchType, secRequirements openapi3.SecurityRequirements) map[string]kuadrantapiv1beta2.AuthenticationSpec {
	// OpenAPI supports as security requirement to have multiple security schemes and ALL
	// of the must be satisfied.
	// From https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#security-requirement-object
	// Kuadrant does not support it yet: https://github.com/Kuadrant/authorino/issues/112
	// not supported (AND'ed)
	// security:
	//   - petstore_api_key: []
	//     petstore_oidc: []
	// supported (OR'ed)
	// security:
	//   - petstore_api_key: []
	//   - petstore_oidc: []

	opAuth := make(map[string]kuadrantapiv1beta2.AuthenticationSpec, 0)
	for _, secReq := range secRequirements {
		if len(secReq) > 1 {
			panic(errors.New("multiple schemes that require ALL must be satisfied, currently not supported"))
		}

		extractSecReqItemName := func(sr openapi3.SecurityRequirement) string {
			for secReqItemName := range sr {
				return secReqItemName
			}

			return ""
		}

		secReqItemName := extractSecReqItemName(secReq)

		secScheme, ok := doc.Components.SecuritySchemes[secReqItemName]
		if !ok {
			// should never happen. OpenAPI validation should detect this issue
			continue
		}

		if secScheme == nil || secScheme.Value == nil {
			continue
		}

		authName := fmt.Sprintf("%s_%s", utils.OpenAPIOperationName(path, verb, op), secReqItemName)

		// Ref https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#fixed-fields-23
		switch secScheme.Value.Type {
		case "openIdConnect":
			opAuth[authName] = openIDAuthenticationSpec(basePath, path, pathItem, verb, op, pathMatchType, *secScheme.Value)
		case "apiKey":
			opAuth[authName] = apiKeyAuthenticationSpec(basePath, path, pathItem, verb, op, pathMatchType, secReqItemName, *secScheme.Value)
		}
	}

	if len(opAuth) == 0 {
		return nil
	}

	return opAuth
}

func apiKeyAuthenticationSpec(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1.PathMatchType, secSchemeName string, secScheme openapi3.SecurityScheme) kuadrantapiv1beta2.AuthenticationSpec {
	// From https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#fixed-fields-23
	// secScheme.In is required
	// secScheme.Name is required
	credentials := authorinoapi.Credentials{}
	switch secScheme.In {
	case "query":
		credentials.QueryString = &authorinoapi.Named{Name: secScheme.Name}
	case "header":
		credentials.CustomHeader = &authorinoapi.CustomHeader{
			Named: authorinoapi.Named{Name: secScheme.Name},
		}
	case "cookie":
		credentials.Cookie = &authorinoapi.Named{Name: secScheme.Name}
	}

	return kuadrantapiv1beta2.AuthenticationSpec{
		CommonAuthRuleSpec: kuadrantapiv1beta2.CommonAuthRuleSpec{
			RouteSelectors: buildAuthPolicyRouteSelectors(basePath, path, pathItem, verb, op, pathMatchType),
		},
		AuthenticationSpec: authorinoapi.AuthenticationSpec{
			Credentials: credentials,
			AuthenticationMethodSpec: authorinoapi.AuthenticationMethodSpec{
				ApiKey: &authorinoapi.ApiKeyAuthenticationSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							// label selector be like
							// kuadrant.io/apikeys-by: ${SecuritySchemeName}
							APIKeySecretLabel: secSchemeName,
						},
					},
				},
			},
		},
	}
}

func openIDAuthenticationSpec(basePath, path string, pathItem *openapi3.PathItem, verb string, op *openapi3.Operation, pathMatchType gatewayapiv1.PathMatchType, secScheme openapi3.SecurityScheme) kuadrantapiv1beta2.AuthenticationSpec {
	return kuadrantapiv1beta2.AuthenticationSpec{
		CommonAuthRuleSpec: kuadrantapiv1beta2.CommonAuthRuleSpec{
			RouteSelectors: buildAuthPolicyRouteSelectors(basePath, path, pathItem, verb, op, pathMatchType),
		},
		AuthenticationSpec: authorinoapi.AuthenticationSpec{
			AuthenticationMethodSpec: authorinoapi.AuthenticationMethodSpec{
				Jwt: &authorinoapi.JwtAuthenticationSpec{
					IssuerUrl: secScheme.OpenIdConnectUrl,
				},
			},
		},
	}
}
