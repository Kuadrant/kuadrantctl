package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	gitSHA  string // value injected in compilation-time
	version string // value injected in compilation-time
)

func versionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kuadrantctl",
		Long:  "Print the version number of kuadrantctl",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("kuadrantctl %s (%s)\n", version, gitSHA)
			return nil
		},
	}
	return versionCmd
}
