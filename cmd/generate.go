package cmd

import (
	"github.com/spf13/cobra"
)

func generateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Commands related to kubernetes object generation",
		Long:  "Commands related to kubernetes object generation",
	}

	cmd.AddCommand(generateIstioCommand())

	return cmd
}
