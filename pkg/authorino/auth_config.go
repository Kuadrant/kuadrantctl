package authorino

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	authorinov1beta1 "github.com/kuadrant/authorino/api/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func AuthConfigIdentitiesFromOpenAPI(oasDoc *openapi3.T) ([]*authorinov1beta1.Identity, error) {
	identities := []*authorinov1beta1.Identity{}

	workloadName, err := utils.K8sNameFromOpenAPITitle(oasDoc)
	if err != nil {
		return nil, err
	}

	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(oasDoc, operation)

			if secReqsP == nil {
				continue
			}

			for _, secReq := range *secReqsP {
				// Authorino AuthConfig currently only supports one identity method for each identity evaluator.
				// It does not support, for instance, auth based on two api keys or api key AND oidc.
				// Thus, some OpenAPI 3.X security requirements are not supported:
				//
				// Not Supported:
				// security:
				//   - petstore_api_key: []
				//     toystore_api_key: []
				//     toystore_oidc: []
				//
				// Supported:
				// security:
				//   - petstore_api_key: []
				//   - toystore_api_key: []
				//   - toystore_oidc: []
				//

				// scopes not being used now
				for secSchemeName := range secReq {

					secSchemeI, err := oasDoc.Components.SecuritySchemes.JSONLookup(secSchemeName)
					if err != nil {
						return nil, err
					}

					secScheme := secSchemeI.(*openapi3.SecurityScheme) // panic if assertion fails

					identity, err := AuthConfigIdentityFromSecurityRequirement(
						operation.OperationID, // TODO(eastizle): OperationID can be null, fallback to some custom name
						path, opVerb, workloadName, secScheme)
					if err != nil {
						return nil, err
					}

					identities = append(identities, identity)
					// currently only support for one schema per requirement
					break
				}
			}

		}
	}
	return identities, nil
}

func AuthConfigConditionsFromOperation(opPath, opVerb string) []authorinov1beta1.JSONPattern {
	return []authorinov1beta1.JSONPattern{
		{
			JSONPatternExpression: authorinov1beta1.JSONPatternExpression{
				Selector: `context.request.http.path@extract{"sep":"/"}`,
				Operator: "eq",
				Value:    opPath,
			},
		},
		{
			JSONPatternExpression: authorinov1beta1.JSONPatternExpression{
				Selector: "context.request.http.method",
				Operator: "eq",
				Value:    opVerb,
			},
		},
	}
}

func AuthConfigIdentityFromSecurityRequirement(name, opPath, opVerb, workloadName string, secScheme *openapi3.SecurityScheme) (*authorinov1beta1.Identity, error) {
	if secScheme == nil {
		return nil, fmt.Errorf("sec scheme nil for operation path:%s method:%s", opPath, opVerb)
	}

	identity := &authorinov1beta1.Identity{
		Name:       name,
		Conditions: AuthConfigConditionsFromOperation(opPath, opVerb),
	}

	switch secScheme.Type {
	case "apiKey":
		AuthConfigIdentityFromApiKeyScheme(identity, secScheme, workloadName)
	case "openIdConnect":
		AuthConfigIdentityFromOIDCScheme(identity, secScheme)
	default:
		return nil, fmt.Errorf("sec scheme type %s not supported for path:%s method:%s", secScheme.Type, opPath, opVerb)
	}

	return identity, nil
}

func AuthConfigIdentityFromApiKeyScheme(identity *authorinov1beta1.Identity, secScheme *openapi3.SecurityScheme, workloadName string) {
	// Fixed label selector for now
	apikey := authorinov1beta1.Identity_APIKey{
		LabelSelectors: map[string]string{
			"authorino.kuadrant.io/managed-by": "authorino",
			"app":                              workloadName,
		},
	}

	identity.Credentials.In = authorinov1beta1.Credentials_In(secScheme.In)
	identity.Credentials.KeySelector = secScheme.Name
	identity.APIKey = &apikey
}

func AuthConfigIdentityFromOIDCScheme(identity *authorinov1beta1.Identity, secScheme *openapi3.SecurityScheme) {
	identity.Oidc = &authorinov1beta1.Identity_OidcConfig{
		Endpoint: secScheme.OpenIdConnectUrl,
	}
}
