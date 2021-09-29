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
	"flag"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	kubeConfig         string
	apiCreateNamespace string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Applies a Kuadrant API, installing on a cluster",
	Long: `The create command generates a Kuadrant API manifest and applies it to a cluster.
For example:

kuadrantctl api create oas3-resource -n ns (/path/to/your/spec/file.[json|yaml|yml] OR
    http[s]://domain/resource/path.[json|yaml|yml] OR '-')
	`,
	Args: cobra.MinimumNArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Required to have controller-runtim config package read the kubeconfig arg
		err := flag.CommandLine.Parse([]string{"-kubeconfig", kubeConfig})
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return apiCreateCmd(cmd, args)
	},
}

func apiCreateCmd(cmd *cobra.Command, args []string) error {
	err := kctlrv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	cl, err := client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	api := &kctlrv1beta1.API{}

	api.SetNamespace(apiCreateNamespace)

	err = cl.Create(context.Background(), api)
	// TODO(eastizle): add type: kind and apiversion
	logf.Log.Info("Created API object", "namespace", apiCreateNamespace, "name", api.GetName(), "error", err)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	// TODO(eastizle): add context flag to switch between kubeconfig contexts
	// It would require using config.GetConfigWithContext(context string) (*rest.Config, error)
	createCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "", "", "Kubernetes configuration file")
	createCmd.Flags().StringVarP(&apiCreateNamespace, "namespace", "n", "", "Cluster namespace (required)")
	err := createCmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}
	apiCmd.AddCommand(createCmd)
}
