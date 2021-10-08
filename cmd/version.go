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

	"github.com/kuadrant/kuadrantctl/pkg/limitador"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
	"github.com/kuadrant/kuadrantctl/version"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func versionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kuadrantctl",
		Long:  "Print the version number of kuadrantctl",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := utils.SetupScheme()
			if err != nil {
				return err
			}

			logf.Log.Info(fmt.Sprintf("kuadrantctl version: %s", version.Version))

			istioVersion, err := utils.IstioImage()
			if err != nil {
				return err
			}
			logf.Log.Info(fmt.Sprintf("Istio version: %s", istioVersion))

			authorinoVersion, err := utils.AuthorinoImage()
			if err != nil {
				return err
			}
			logf.Log.Info(fmt.Sprintf("Authorino version: %s", authorinoVersion))

			limitadorOperatorVersion, err := utils.LimitadorOperatorImage()
			if err != nil {
				return err
			}
			logf.Log.Info(fmt.Sprintf("Limitador operator version: %s", limitadorOperatorVersion))

			limitadorObj := limitador.Limitador(installNamespace)
			logf.Log.Info(fmt.Sprintf("Limitador version: %s", *limitadorObj.Spec.Version))

			kuadrantControllerVersion, err := utils.KuadrantControllerImage()
			if err != nil {
				return err
			}
			logf.Log.Info(fmt.Sprintf("Kuadrant controller version: %s", kuadrantControllerVersion))

			return nil
		},
	}
	return versionCmd
}
