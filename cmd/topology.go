package cmd

import (
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	topologyNS         string
	topologyOutputFile string
)

func topologyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topology",
		Short: "Read kuadrant topology",
		Long:  "Read kuadrant topology",
		RunE:  runTopology,
	}

	cmd.Flags().StringVarP(&topologyNS, "namespace", "n", "kuadrant-system", "Topology's namespace")
	cmd.Flags().StringVarP(&topologyOutputFile, "output", "o", "/dev/stdout", "Output file")
	return cmd
}

func runTopology(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	k8sClient, err := client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	topologyKey := client.ObjectKey{Name: "topology", Namespace: topologyNS}
	topologyConfigMap := &corev1.ConfigMap{}
	err = k8sClient.Get(ctx, topologyKey, topologyConfigMap)
	logf.Log.V(1).Info("reading topology", "object", client.ObjectKeyFromObject(topologyConfigMap), "error", err)
	if err != nil {
		return err
	}
	f, err := os.Create(topologyOutputFile)
	logf.Log.V(1).Info("write topology topology to file", "file", topologyOutputFile, "error", err)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(topologyConfigMap.Data["topology"])
	if err != nil {
		return err
	}

	return nil
}
