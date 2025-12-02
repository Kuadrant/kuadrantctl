/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	kuadrantv1 "github.com/kuadrant/kuadrant-operator/api/v1"
	kuadrantv1alpha1 "github.com/kuadrant/kuadrant-operator/api/v1alpha1"
	kuadrantv1beta1 "github.com/kuadrant/kuadrant-operator/api/v1beta1"
)

var (
	verbose bool
)

// newK8sClient creates a new Kubernetes client with Gateway API and Kuadrant schemes registered
func newK8sClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// Register Gateway API and Kuadrant schemes
	schemesToRegister := []struct {
		addFunc func(*runtime.Scheme) error
		name    string
	}{
		{gatewayapiv1.AddToScheme, "Gateway API"},
		{kuadrantv1.AddToScheme, "Kuadrant v1 API"},
		{kuadrantv1alpha1.AddToScheme, "Kuadrant v1alpha1 API"},
		{kuadrantv1beta1.AddToScheme, "Kuadrant v1beta1 API"},
	}

	for _, s := range schemesToRegister {
		if err := s.addFunc(scheme.Scheme); err != nil {
			return nil, fmt.Errorf("failed to add %s to scheme: %w", s.name, err)
		}
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return k8sClient, nil
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "kuadrantctl",
		Short: "Kuadrant configuration command line utility",
		Long:  "Kuadrant configuration command line utility",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logf.SetLogger(zap.New(zap.UseDevMode(verbose), zap.WriteTo(os.Stdout)))
			cmd.SetContext(context.Background())
		},
	}

	rootCmd.SetArgs(args)

	// avoid usage being shown on error
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(generateCommand())
	rootCmd.AddCommand(topologyCommand())
	rootCmd.AddCommand(dumpCommand())
	rootCmd.AddCommand(diagnoseCommand())

	if isBinaryAvailable("kubectl-kuadrant_dns") {
		rootCmd.AddCommand(dnsCommand())
	}

	return rootCmd
}

func isBinaryAvailable(name string) bool {
	path, err := exec.LookPath(name)
	return path != "" && err == nil
}
