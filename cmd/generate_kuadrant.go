package cmd

import (
	"github.com/spf13/cobra"
)

func generateKuadrantCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kuadrant",
		Short: "Generate Kuadrant resources",
		Long:  "Generate Kuadrant resources",
	}

	cmd.AddCommand(generateKuadrantAuthconfigCommand())

	return cmd
}
