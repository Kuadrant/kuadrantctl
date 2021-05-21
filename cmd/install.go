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
	"fmt"

	"github.com/spf13/cobra"
)

var (
	installContext    string
	installKubeConfig string
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Applies an kuadrant manifest, installing or reconfiguring kuadrant on a cluster",
	Long:  "The install command generates an Istio install manifest and applies it to a cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return installRun(cmd, args)
	},
}

func installRun(cmd *cobra.Command, args []string) error {
	fmt.Println("install called")
	return nil
}

func init() {
	installCmd.PersistentFlags().StringVarP(&installContext, "context", "", "", "The name of the kubeconfig context to use")
	installCmd.PersistentFlags().StringVarP(&installKubeConfig, "kubeconfig", "c", "", "Kubernetes configuration file")
	rootCmd.AddCommand(installCmd)
}
