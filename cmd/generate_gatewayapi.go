package cmd

import (
	"github.com/spf13/cobra"
)

func generateGatewayAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gatewayapi",
		Short: "Generate Gataway API resources",
		Long:  "Generate Gataway API resources",
	}

	cmd.AddCommand(generateGatewayApiHttpRouteCommand())

	return cmd
}
