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
	"github.com/spf13/cobra"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	verbose bool
)

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "kuadrantctl",
		Short: "Kuadrant configuration command line utility",
		Long:  "Kuadrant configuration command line utility",
	}

	rootCmd.SetArgs(args)

	// avoid usage being shown on error
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(apiCommand())
	rootCmd.AddCommand(installCommand())
	rootCmd.AddCommand(uninstallCommand())
	rootCmd.AddCommand(versionCommand())

	loggerOpts := zap.Options{Development: verbose}
	logf.SetLogger(zap.New(zap.UseFlagOptions(&loggerOpts)))

	return rootCmd
}
