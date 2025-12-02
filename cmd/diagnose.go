package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	kuadrantv1 "github.com/kuadrant/kuadrant-operator/api/v1"
	kuadrantv1beta1 "github.com/kuadrant/kuadrant-operator/api/v1beta1"
)

var (
	diagnoseNamespace     string
	diagnoseAllNamespaces bool
	diagnoseOutputFormat  string
)

type DiagnosticReport struct {
	GatewayIssues         []ResourceIssue
	HTTPRouteIssues       []ResourceIssue
	AuthPolicyIssues      []ResourceIssue
	RateLimitPolicyIssues []ResourceIssue
	DNSPolicyIssues       []ResourceIssue
	TLSPolicyIssues       []ResourceIssue
	KuadrantIssues        []ResourceIssue
	Summary               DiagnosticSummary
}

type ResourceIssue struct {
	ResourceType string
	Namespace    string
	Name         string
	Issues       []string
	Status       string
}

type DiagnosticSummary struct {
	TotalGateways               int
	UnprogrammedGateways        int
	TotalHTTPRoutes             int
	UnacceptedHTTPRoutes        int
	TotalAuthPolicies           int
	UnenforcedAuthPolicies      int
	TotalRateLimitPolicies      int
	UnenforcedRateLimitPolicies int
	TotalDNSPolicies            int
	UnenforcedDNSPolicies       int
	TotalTLSPolicies            int
	UnenforcedTLSPolicies       int
	TotalKuadrants              int
	UnreadyKuadrants            int
}

// checkPolicyConditions checks standard Accepted and Enforced conditions on a policy
// Returns a list of issues and whether the policy is enforced
func checkPolicyConditions(conditions []metav1.Condition) (issues []string, isEnforced bool) {
	for _, condition := range conditions {
		switch condition.Type {
		case "Accepted":
			if string(condition.Status) != string(metav1.ConditionTrue) {
				issues = append(issues, fmt.Sprintf("Not Accepted: %s - %s", condition.Reason, condition.Message))
			}
		case "Enforced":
			if string(condition.Status) != string(metav1.ConditionTrue) {
				issues = append(issues, fmt.Sprintf("Not Enforced: %s - %s", condition.Reason, condition.Message))
			} else {
				isEnforced = true
			}
		}
	}
	return issues, isEnforced
}

func diagnoseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Diagnose Kuadrant and Gateway API resources",
		Long:  "Analyze Kuadrant and Gateway API resources for issues, including unprogrammed gateways, unenforced policies, and configuration problems",
		RunE:  runDiagnose,
	}

	cmd.Flags().StringVarP(&diagnoseNamespace, "namespace", "n", "", "Namespace to diagnose (default: all namespaces)")
	cmd.Flags().BoolVarP(&diagnoseAllNamespaces, "all-namespaces", "A", false, "Diagnose resources from all namespaces")
	cmd.Flags().StringVarP(&diagnoseOutputFormat, "output", "o", "text", "Output format: text, yaml, or json")

	return cmd
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	logger := logf.FromContext(ctx)

	k8sClient, err := newK8sClient()
	if err != nil {
		return err
	}

	// Determine namespace scope
	var namespaceOption client.ListOption
	if diagnoseAllNamespaces || diagnoseNamespace == "" {
		namespaceOption = client.InNamespace("")
		logger.Info("Diagnosing resources from all namespaces")
	} else {
		namespaceOption = client.InNamespace(diagnoseNamespace)
		logger.Info("Diagnosing resources", "namespace", diagnoseNamespace)
	}

	report := DiagnosticReport{}

	// Define all diagnose functions to run
	diagnoseFuncs := []struct {
		name string
		fn   func(context.Context, client.Client, *DiagnosticReport, client.ListOption) error
	}{
		{"gateways", diagnoseGateways},
		{"HTTPRoutes", diagnoseHTTPRoutes},
		{"AuthPolicies", diagnoseAuthPolicies},
		{"RateLimitPolicies", diagnoseRateLimitPolicies},
		{"DNSPolicies", diagnoseDNSPolicies},
		{"TLSPolicies", diagnoseTLSPolicies},
		{"Kuadrant CRs", diagnoseKuadrants},
	}

	// Run all diagnose functions
	for _, df := range diagnoseFuncs {
		if err := df.fn(ctx, k8sClient, &report, namespaceOption); err != nil {
			logger.Error(err, "Failed to diagnose", "type", df.name)
			// Continue with other resource types
		}
	}

	// Print report
	printDiagnosticReport(report)

	return nil
}

func diagnoseGateways(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	gatewayList := &gatewayapiv1.GatewayList{}
	if err := k8sClient.List(ctx, gatewayList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalGateways = len(gatewayList.Items)

	for _, gateway := range gatewayList.Items {
		issues := []string{}
		isProgrammed := false
		isAccepted := false

		// Check status conditions
		for _, condition := range gateway.Status.Conditions {
			switch condition.Type {
			case string(gatewayapiv1.GatewayConditionProgrammed):
				if string(condition.Status) != string(metav1.ConditionTrue) {
					issues = append(issues, fmt.Sprintf("Not Programmed: %s - %s", condition.Reason, condition.Message))
				} else {
					isProgrammed = true
				}
			case string(gatewayapiv1.GatewayConditionAccepted):
				if string(condition.Status) != string(metav1.ConditionTrue) {
					issues = append(issues, fmt.Sprintf("Not Accepted: %s - %s", condition.Reason, condition.Message))
				} else {
					isAccepted = true
				}
			case string(gatewayapiv1.GatewayConditionReady):
				if string(condition.Status) != string(metav1.ConditionTrue) {
					issues = append(issues, fmt.Sprintf("Not Ready: %s - %s", condition.Reason, condition.Message))
				}
			}
		}

		// Check listener statuses
		// Only check positive conditions (conditions that should be True for healthy state)
		for _, listener := range gateway.Status.Listeners {
			for _, condition := range listener.Conditions {
				// Only report issues for conditions that should be True
				// Negative conditions like "Conflicted" or "Detached" being False is healthy
				switch condition.Type {
				case string(gatewayapiv1.ListenerConditionAccepted),
					string(gatewayapiv1.ListenerConditionProgrammed),
					string(gatewayapiv1.ListenerConditionResolvedRefs):
					if string(condition.Status) != string(metav1.ConditionTrue) {
						issues = append(issues, fmt.Sprintf("Listener %s - %s: %s - %s",
							listener.Name, condition.Type, condition.Reason, condition.Message))
					}
				}
			}
		}

		// Count unprogrammed gateways once per gateway
		if !isProgrammed {
			report.Summary.UnprogrammedGateways++
		}

		if len(issues) > 0 {
			var status string
			if isProgrammed && isAccepted {
				status = "Programmed+Accepted (with warnings)"
			} else if !isProgrammed {
				status = "Not Programmed"
			} else if !isAccepted {
				status = "Not Accepted"
			} else {
				status = "Unknown"
			}

			report.GatewayIssues = append(report.GatewayIssues, ResourceIssue{
				ResourceType: "Gateway",
				Namespace:    gateway.Namespace,
				Name:         gateway.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseHTTPRoutes(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	routeList := &gatewayapiv1.HTTPRouteList{}
	if err := k8sClient.List(ctx, routeList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalHTTPRoutes = len(routeList.Items)

	for _, route := range routeList.Items {
		issues := []string{}
		isAccepted := false

		// Check parent status
		for _, parentStatus := range route.Status.Parents {
			for _, condition := range parentStatus.Conditions {
				if condition.Type == string(gatewayapiv1.RouteConditionAccepted) {
					if string(condition.Status) != string(metav1.ConditionTrue) {
						issues = append(issues, fmt.Sprintf("Not Accepted by parent %s: %s - %s",
							parentStatus.ParentRef.Name, condition.Reason, condition.Message))
					} else {
						isAccepted = true
					}
				} else if string(condition.Status) != string(metav1.ConditionTrue) {
					issues = append(issues, fmt.Sprintf("Parent %s - %s: %s - %s",
						parentStatus.ParentRef.Name, condition.Type, condition.Reason, condition.Message))
				}
			}
		}

		// Check if route has any parent refs
		if len(route.Spec.ParentRefs) == 0 {
			issues = append(issues, "No parent gateways configured")
		}

		// Count unaccepted routes once per route
		if !isAccepted && len(issues) > 0 {
			report.Summary.UnacceptedHTTPRoutes++
		}

		if len(issues) > 0 {
			var status string
			if isAccepted {
				status = "Accepted (with warnings)"
			} else {
				status = "Not Accepted"
			}

			report.HTTPRouteIssues = append(report.HTTPRouteIssues, ResourceIssue{
				ResourceType: "HTTPRoute",
				Namespace:    route.Namespace,
				Name:         route.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseAuthPolicies(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	policyList := &kuadrantv1.AuthPolicyList{}
	if err := k8sClient.List(ctx, policyList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalAuthPolicies = len(policyList.Items)

	for _, policy := range policyList.Items {
		// Check status conditions
		conditionIssues, isEnforced := checkPolicyConditions(policy.Status.Conditions)
		issues := conditionIssues

		if !isEnforced {
			report.Summary.UnenforcedAuthPolicies++
		}

		// Check if target ref is set
		if policy.Spec.TargetRef.Name == "" {
			issues = append(issues, "No target reference configured")
		}

		if len(issues) > 0 {
			var status string
			if isEnforced {
				status = "Enforced (with warnings)"
			} else {
				status = "Not Enforced"
			}

			report.AuthPolicyIssues = append(report.AuthPolicyIssues, ResourceIssue{
				ResourceType: "AuthPolicy",
				Namespace:    policy.Namespace,
				Name:         policy.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseRateLimitPolicies(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	policyList := &kuadrantv1.RateLimitPolicyList{}
	if err := k8sClient.List(ctx, policyList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalRateLimitPolicies = len(policyList.Items)

	for _, policy := range policyList.Items {
		// Check status conditions
		conditionIssues, isEnforced := checkPolicyConditions(policy.Status.Conditions)
		issues := conditionIssues

		if !isEnforced {
			report.Summary.UnenforcedRateLimitPolicies++
		}

		// Check if target ref is set
		if policy.Spec.TargetRef.Name == "" {
			issues = append(issues, "No target reference configured")
		}

		if len(issues) > 0 {
			var status string
			if isEnforced {
				status = "Enforced (with warnings)"
			} else {
				status = "Not Enforced"
			}

			report.RateLimitPolicyIssues = append(report.RateLimitPolicyIssues, ResourceIssue{
				ResourceType: "RateLimitPolicy",
				Namespace:    policy.Namespace,
				Name:         policy.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseDNSPolicies(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	policyList := &kuadrantv1.DNSPolicyList{}
	if err := k8sClient.List(ctx, policyList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalDNSPolicies = len(policyList.Items)

	for _, policy := range policyList.Items {
		// Check status conditions
		conditionIssues, isEnforced := checkPolicyConditions(policy.Status.Conditions)
		issues := conditionIssues

		if !isEnforced {
			report.Summary.UnenforcedDNSPolicies++
		}

		// Check if target ref is set
		if policy.Spec.TargetRef.Name == "" {
			issues = append(issues, "No target reference configured")
		}

		if len(issues) > 0 {
			var status string
			if isEnforced {
				status = "Enforced (with warnings)"
			} else {
				status = "Not Enforced"
			}

			report.DNSPolicyIssues = append(report.DNSPolicyIssues, ResourceIssue{
				ResourceType: "DNSPolicy",
				Namespace:    policy.Namespace,
				Name:         policy.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseTLSPolicies(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	policyList := &kuadrantv1.TLSPolicyList{}
	if err := k8sClient.List(ctx, policyList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalTLSPolicies = len(policyList.Items)

	for _, policy := range policyList.Items {
		// Check status conditions
		conditionIssues, isEnforced := checkPolicyConditions(policy.Status.Conditions)
		issues := conditionIssues

		if !isEnforced {
			report.Summary.UnenforcedTLSPolicies++
		}

		// Check if target ref is set
		if policy.Spec.TargetRef.Name == "" {
			issues = append(issues, "No target reference configured")
		}

		if len(issues) > 0 {
			var status string
			if isEnforced {
				status = "Enforced (with warnings)"
			} else {
				status = "Not Enforced"
			}

			report.TLSPolicyIssues = append(report.TLSPolicyIssues, ResourceIssue{
				ResourceType: "TLSPolicy",
				Namespace:    policy.Namespace,
				Name:         policy.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func diagnoseKuadrants(ctx context.Context, k8sClient client.Client, report *DiagnosticReport, namespaceOption client.ListOption) error {
	kuadrantList := &kuadrantv1beta1.KuadrantList{}
	if err := k8sClient.List(ctx, kuadrantList, namespaceOption); err != nil {
		return err
	}

	report.Summary.TotalKuadrants = len(kuadrantList.Items)

	for _, kuadrant := range kuadrantList.Items {
		issues := []string{}
		isReady := false

		// Check status conditions
		for _, condition := range kuadrant.Status.Conditions {
			if string(condition.Status) != string(metav1.ConditionTrue) {
				issues = append(issues, fmt.Sprintf("%s: %s - %s", condition.Type, condition.Reason, condition.Message))
				if condition.Type == "Ready" {
					report.Summary.UnreadyKuadrants++
				}
			} else if condition.Type == "Ready" {
				isReady = true
			}
		}

		if len(issues) > 0 {
			var status string
			if isReady {
				status = "Ready (with warnings)"
			} else {
				status = "Not Ready"
			}

			report.KuadrantIssues = append(report.KuadrantIssues, ResourceIssue{
				ResourceType: "Kuadrant",
				Namespace:    kuadrant.Namespace,
				Name:         kuadrant.Name,
				Issues:       issues,
				Status:       status,
			})
		}
	}

	return nil
}

func printDiagnosticReport(report DiagnosticReport) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("KUADRANT DIAGNOSTIC REPORT")
	fmt.Println(strings.Repeat("=", 80))

	// Print summary
	fmt.Println("\nSUMMARY:")
	fmt.Println(strings.Repeat("-", 80))

	summaryItems := []struct {
		label      string
		total      int
		unhealthy  int
		statusWord string
	}{
		{"Gateways", report.Summary.TotalGateways, report.Summary.UnprogrammedGateways, "unprogrammed"},
		{"HTTPRoutes", report.Summary.TotalHTTPRoutes, report.Summary.UnacceptedHTTPRoutes, "unaccepted"},
		{"AuthPolicies", report.Summary.TotalAuthPolicies, report.Summary.UnenforcedAuthPolicies, "unenforced"},
		{"RateLimitPolicies", report.Summary.TotalRateLimitPolicies, report.Summary.UnenforcedRateLimitPolicies, "unenforced"},
		{"DNSPolicies", report.Summary.TotalDNSPolicies, report.Summary.UnenforcedDNSPolicies, "unenforced"},
		{"TLSPolicies", report.Summary.TotalTLSPolicies, report.Summary.UnenforcedTLSPolicies, "unenforced"},
		{"Kuadrants", report.Summary.TotalKuadrants, report.Summary.UnreadyKuadrants, "unready"},
	}

	for _, item := range summaryItems {
		fmt.Printf("%-27s %d total, %d %s\n", item.label+":", item.total, item.unhealthy, item.statusWord)
	}

	// Count total issues
	totalIssues := len(report.GatewayIssues) + len(report.HTTPRouteIssues) +
		len(report.AuthPolicyIssues) + len(report.RateLimitPolicyIssues) +
		len(report.DNSPolicyIssues) + len(report.TLSPolicyIssues) +
		len(report.KuadrantIssues)

	if totalIssues == 0 {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("✓ No issues found! All resources are healthy.")
		fmt.Println(strings.Repeat("=", 80))
		return
	}

	fmt.Printf("\nTotal issues found: %d\n", totalIssues)

	// Print all resource issues
	issueGroups := []struct {
		title  string
		issues []ResourceIssue
	}{
		{"GATEWAY ISSUES:", report.GatewayIssues},
		{"HTTPROUTE ISSUES:", report.HTTPRouteIssues},
		{"AUTHPOLICY ISSUES:", report.AuthPolicyIssues},
		{"RATELIMITPOLICY ISSUES:", report.RateLimitPolicyIssues},
		{"DNSPOLICY ISSUES:", report.DNSPolicyIssues},
		{"TLSPOLICY ISSUES:", report.TLSPolicyIssues},
		{"KUADRANT ISSUES:", report.KuadrantIssues},
	}

	for _, group := range issueGroups {
		if len(group.issues) > 0 {
			fmt.Println("\n" + strings.Repeat("-", 80))
			fmt.Println(group.title)
			fmt.Println(strings.Repeat("-", 80))
			for _, issue := range group.issues {
				printResourceIssue(issue)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

func printResourceIssue(issue ResourceIssue) {
	fmt.Printf("\n  Resource: %s/%s\n", issue.Namespace, issue.Name)
	fmt.Printf("  Status:   %s\n", issue.Status)
	fmt.Println("  Issues:")
	for _, iss := range issue.Issues {
		fmt.Printf("    - %s\n", iss)
	}
}
