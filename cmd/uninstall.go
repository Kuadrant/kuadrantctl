package cmd

import (
	"flag"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kuadrantoperator "github.com/kuadrant/kuadrant-operator/api/v1beta1"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

func uninstallCommand() *cobra.Command {
	unInstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalling kuadrant from the cluster",
		Long:  "The uninstall command removes kuadrant manifest bundle from the cluster.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Required to have controller-runtim config package read the kubeconfig arg
			err := flag.CommandLine.Parse([]string{"-kubeconfig", installKubeConfig})
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return unInstallRun(cmd, args)
		},
	}

	// TODO(eastizle): add context flag to switch between kubeconfig contexts
	// It would require using config.GetConfigWithContext(context string) (*rest.Config, error)
	unInstallCmd.PersistentFlags().StringVarP(&installKubeConfig, "kubeconfig", "", "", "Kubernetes configuration file")

	return unInstallCmd
}

func unInstallRun(cmd *cobra.Command, args []string) error {
	err := utils.SetupScheme()
	if err != nil {
		return err
	}

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	k8sClient, err := client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	err = unDeployKuadrant(k8sClient)
	if err != nil {
		return err
	}

	logf.Log.Info("kuadrant successfully uninstalled")

	return nil
}

func unDeployKuadrant(k8sClient client.Client) error {
	kuadrant := &kuadrantoperator.Kuadrant{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kuadrant.io/v1beta1", Kind: "Kuadrant"},
		ObjectMeta: metav1.ObjectMeta{Name: "kuadrant", Namespace: installNamespace},
	}

	err := utils.DeleteK8SObject(k8sClient, kuadrant)
	if err != nil {
		return err
	}

	return nil
}
