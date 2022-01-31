package istio

import (
	"github.com/getkin/kin-openapi/openapi3"
	istiosecurityapi "istio.io/api/security/v1beta1"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func AuthorizationPolicyRulesFromOpenAPI(oasDoc *openapi3.T, publicDomain string) []*istiosecurityapi.Rule {
	rules := []*istiosecurityapi.Rule{}

	for path, pathItem := range oasDoc.Paths {
		for opVerb, operation := range pathItem.Operations() {
			secReqsP := utils.OpenAPIOperationSecRequirements(oasDoc, operation)

			if secReqsP == nil || len(*secReqsP) == 0 {
				continue
			}

			// there is at least one sec requirement for this operation,
			// add the operation to authorization policy rules
			rule := &istiosecurityapi.Rule{
				To: []*istiosecurityapi.Rule_To{
					{
						Operation: &istiosecurityapi.Operation{
							Hosts:   []string{publicDomain},
							Methods: []string{opVerb},
							Paths:   []string{path},
						},
					},
				},
			}

			rules = append(rules, rule)
		}
	}
	return rules
}
