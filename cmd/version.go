package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuadrant/kuadrantctl/pkg/authorino"
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

			authorinoOperatorVersion, err := utils.AuthorinoOperatorImage()
			if err != nil {
				return err
			}
			logf.Log.Info(fmt.Sprintf("Authorino operator version: %s", authorinoOperatorVersion))

			authorinoObj := authorino.Authorino(installNamespace)
			logf.Log.Info(fmt.Sprintf("Authorino version: %s", authorinoObj.Spec.Image))

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
