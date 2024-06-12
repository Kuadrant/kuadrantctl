package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
	"github.com/kuadrant/kuadrantctl/version"
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

			fmt.Printf("kuadrantctl %s (%s)\n", version.Version, version.GitHash)
			return nil
		},
	}
	return versionCmd
}
