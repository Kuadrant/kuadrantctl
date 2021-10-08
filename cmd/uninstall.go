package cmd

import (
	"flag"
	"reflect"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kuadrant/kuadrantctl/authorinomanifests"
	"github.com/kuadrant/kuadrantctl/istiomanifests"
	"github.com/kuadrant/kuadrantctl/kuadrantmanifests"
	"github.com/kuadrant/kuadrantctl/limitadormanifests"
	"github.com/kuadrant/kuadrantctl/pkg/limitador"
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

	err = unDeployAuthorizationProvider(k8sClient)
	if err != nil {
		return err
	}

	err = unDeployIngressProvider(k8sClient)
	if err != nil {
		return err
	}

	err = unDeployRateLimitProvider(k8sClient)
	if err != nil {
		return err
	}

	logf.Log.Info("kuadrant successfully removed")

	return nil
}

func unDeployKuadrant(k8sClient client.Client) error {
	data, err := kuadrantmanifests.Content()
	if err != nil {
		return err
	}

	if err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient)); err != nil {
		return err
	}
	return nil
}

func unDeployAuthorizationProvider(k8sClient client.Client) error {
	data, err := authorinomanifests.Content()
	if err != nil {
		return err
	}

	err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient))
	if err != nil {
		return err
	}

	return nil
}

func unDeployIngressProvider(k8sClient client.Client) error {
	manifests := []struct {
		source func() ([]byte, error)
	}{
		{istiomanifests.BaseContent},
		{istiomanifests.PilotContent},
		{istiomanifests.IngressGatewayContent},
		{istiomanifests.DefaultGatewayContent},
	}

	for _, manifest := range manifests {
		data, err := manifest.source()
		if err != nil {
			return err
		}
		err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient))
		if err != nil {
			return err
		}
	}

	return nil
}

func unDeployRateLimitProvider(k8sClient client.Client) error {
	err := utils.DeleteK8SObject(k8sClient, limitador.Limitador(installNamespace))
	if err != nil {
		return err
	}

	data, err := limitadormanifests.OperatorContent()
	if err != nil {
		return err
	}
	err = utils.DecodeFile(data, scheme.Scheme, delete(k8sClient))
	if err != nil {
		return err
	}

	return nil
}

func delete(k8sClient client.Client) utils.DecodeCallback {
	return func(obj runtime.Object) error {
		if (obj.GetObjectKind().GroupVersionKind().GroupVersion() == corev1.SchemeGroupVersion && obj.GetObjectKind().GroupVersionKind().Kind == reflect.TypeOf(corev1.Namespace{}).Name()) ||
			obj.GetObjectKind().GroupVersionKind().Group == apiextensionsv1beta1.GroupName || obj.GetObjectKind().GroupVersionKind().Group == apiextensionsv1.GroupName {
			// Omit Namespace and CRD's deletion inside the manifest data
			return nil
		}

		return utils.DeleteK8SObject(k8sClient, obj)
	}
}
