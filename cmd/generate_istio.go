package cmd

import (
	"github.com/spf13/cobra"
)

func generateIstioCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio",
		Short: "Generate Istio resources",
		Long:  "Generate Istio resorces",
	}

	cmd.AddCommand(generateIstioVirtualServiceCommand())
	cmd.AddCommand(generateIstioAuthorizationPolicyCommand())

	return cmd
}
