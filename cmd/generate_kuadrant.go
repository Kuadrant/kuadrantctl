package cmd

import (
	"github.com/spf13/cobra"
)

func generateKuadrantCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kuadrant",
		Short: "Generate Kuadrant resources",
		Long:  "Generate Kuadrant resorces",
	}

	cmd.AddCommand(generateKuadrantAuthconfigCommand())

	return cmd
}
