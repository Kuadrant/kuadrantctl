package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func dnsCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "dns",
		Short: "DNS Operator command line utility",
		Long:  "DNS Operator command line utility",
		RunE:  runDNS,
	}
	return versionCmd
}

func runDNS(_ *cobra.Command, args []string) error {
	// pass verbose from root
	args = append(args, fmt.Sprintf("--verbose=%t", verbose))

	out, err := exec.Command("kubectl-dns", args...).Output()
	if err != nil {
		return fmt.Errorf("failed to run dns plugin: %w", err)
	}
	fmt.Printf("%s\n", out)
	return nil
}
