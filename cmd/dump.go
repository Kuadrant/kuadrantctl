package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	kuadrantv1 "github.com/kuadrant/kuadrant-operator/api/v1"
	kuadrantv1beta1 "github.com/kuadrant/kuadrant-operator/api/v1beta1"
)

var (
	dumpNamespace     string
	dumpAllNamespaces bool
	dumpOutputDir     string
	dumpForce         bool
)

func dumpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dump Kuadrant and Gateway API resources",
		Long:  "Dump all Kuadrant and Gateway API resources to files for investigation and debugging",
		RunE:  runDump,
	}

	cmd.Flags().StringVarP(&dumpNamespace, "namespace", "n", "", "Namespace to dump resources from")
	cmd.Flags().BoolVarP(&dumpAllNamespaces, "all-namespaces", "A", false, "Explicitly dump resources from all namespaces (default if no namespace specified)")
	cmd.Flags().StringVarP(&dumpOutputDir, "output", "o", "", "Output directory (default: ./kuadrant-dump-<timestamp>)")
	cmd.Flags().BoolVar(&dumpForce, "force", false, "Force overwrite if output directory exists and is not empty")

	return cmd
}

func runDump(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	logger := logf.FromContext(ctx)

	// Validate flags
	if dumpNamespace != "" && dumpAllNamespaces {
		return fmt.Errorf("cannot specify both --namespace and --all-namespaces")
	}

	if dumpOutputDir == "" {
		timestamp := time.Now().Format("20060102-150405")
		dumpOutputDir = fmt.Sprintf("kuadrant-dump-%s", timestamp)
	}

	// Check if directory exists and is not empty (unless --force is used)
	if !dumpForce {
		if info, err := os.Stat(dumpOutputDir); err == nil && info.IsDir() {
			entries, err := os.ReadDir(dumpOutputDir)
			if err != nil {
				return fmt.Errorf("failed to read output directory: %w", err)
			}
			if len(entries) > 0 {
				return fmt.Errorf("output directory %s already exists and is not empty (use --force to overwrite)", dumpOutputDir)
			}
		}
	}

	if err := os.MkdirAll(dumpOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	logger.Info("Dumping resources", "output", dumpOutputDir)

	k8sClient, err := newK8sClient()
	if err != nil {
		return err
	}

	// Determine namespace scope
	var namespaceOption client.ListOption
	if dumpAllNamespaces || dumpNamespace == "" {
		namespaceOption = client.InNamespace("")
		logger.Info("Dumping resources from all namespaces")
	} else {
		namespaceOption = client.InNamespace(dumpNamespace)
		logger.Info("Dumping resources", "namespace", dumpNamespace)
	}

	// Dump resources
	resourceTypes := []struct {
		name    string
		listObj client.ObjectList
	}{
		{"gateways", &gatewayapiv1.GatewayList{}},
		{"gatewayclasses", &gatewayapiv1.GatewayClassList{}},
		{"httproutes", &gatewayapiv1.HTTPRouteList{}},
		{"authpolicies", &kuadrantv1.AuthPolicyList{}},
		{"ratelimitpolicies", &kuadrantv1.RateLimitPolicyList{}},
		{"dnspolicies", &kuadrantv1.DNSPolicyList{}},
		{"tlspolicies", &kuadrantv1.TLSPolicyList{}},
		{"kuadrants", &kuadrantv1beta1.KuadrantList{}},
	}

	for _, rt := range resourceTypes {
		if err := dumpResourceTypeGeneric(ctx, k8sClient, rt.name, rt.listObj, namespaceOption); err != nil {
			logger.Error(err, "Failed to dump resource type", "type", rt.name)
			// Continue with other resource types
		}
	}

	logger.Info("Dump completed successfully", "output", dumpOutputDir)
	fmt.Printf("\nResources dumped to: %s\n", dumpOutputDir)

	return nil
}

func dumpResourceTypeGeneric(
	ctx context.Context,
	k8sClient client.Client,
	resourceType string,
	listObj client.ObjectList,
	namespaceOption client.ListOption,
) error {
	logger := logf.FromContext(ctx)

	if err := k8sClient.List(ctx, listObj, namespaceOption); err != nil {
		return fmt.Errorf("failed to list %s: %w", resourceType, err)
	}

	listValue := reflect.ValueOf(listObj).Elem()
	itemsField := listValue.FieldByName("Items")
	if !itemsField.IsValid() {
		return fmt.Errorf("list type for %s does not have Items field", resourceType)
	}

	itemsLen := itemsField.Len()
	if itemsLen == 0 {
		logger.V(1).Info("No resources found", "type", resourceType)
		return nil
	}

	logger.Info("Dumping resources", "type", resourceType, "count", itemsLen)

	for i := 0; i < itemsLen; i++ {
		item := itemsField.Index(i).Addr().Interface()
		obj := item.(client.Object)

		filename := fmt.Sprintf("%s-%s.yaml", obj.GetNamespace(), obj.GetName())
		if obj.GetNamespace() == "" {
			filename = fmt.Sprintf("%s.yaml", obj.GetName())
		}

		filePath := filepath.Join(dumpOutputDir, resourceType, filename)

		data, err := yaml.Marshal(item)
		if err != nil {
			logger.Error(err, "Failed to marshal resource", "name", obj.GetName(), "namespace", obj.GetNamespace())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			logger.Error(err, "Failed to create directory", "path", filepath.Dir(filePath))
			continue
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			logger.Error(err, "Failed to write file", "file", filePath)
			continue
		}
	}

	return nil
}
