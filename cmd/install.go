package cmd

import (
	"context"
	"flag"
	"fmt"
	"time"

	operators "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"

	kuadrantoperator "github.com/kuadrant/kuadrant-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	installKubeConfig string
	// TODO(eastizle): namespace from command line param
	installNamespace string = "kuadrant-system"
)

const (
	KUADRANT_OPERATOR_VERSION string = "v0.4.1"
)

func installCommand() *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Applies a kuadrant manifest bundle, installing or reconfiguring kuadrant on a cluster",
		Long:  "The install command applies kuadrant manifest bundle and applies it to a cluster.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Required to have controller-runtim config package read the kubeconfig arg
			err := flag.CommandLine.Parse([]string{"-kubeconfig", installKubeConfig})
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return installRun(cmd, args)
		},
	}

	// TODO(eastizle): add context flag to switch between kubeconfig contexts
	// It would require using config.GetConfigWithContext(context string) (*rest.Config, error)
	installCmd.PersistentFlags().StringVarP(&installKubeConfig, "kubeconfig", "", "", "Kubernetes configuration file")
	return installCmd
}

func installRun(cmd *cobra.Command, args []string) error {
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

	err = deployKuadrantOperator(k8sClient)
	if err != nil {
		return err
	}

	err = createNamespace(k8sClient)
	if err != nil {
		return err
	}

	err = deployKuadrant(k8sClient)
	if err != nil {
		return err
	}

	err = waitForDeployments(k8sClient)
	if err != nil {
		return err
	}

	return nil
}

func waitForDeployments(k8sClient client.Client) error {
	retryInterval := time.Second * 5
	timeout := time.Minute * 2

	deploymentKeys := []types.NamespacedName{
		types.NamespacedName{Name: "authorino", Namespace: installNamespace},
		types.NamespacedName{Name: "limitador", Namespace: installNamespace},
	}

	for _, key := range deploymentKeys {
		err := wait.Poll(retryInterval, timeout, func() (bool, error) {
			return utils.CheckDeploymentAvailable(k8sClient, key)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func deployKuadrantOperator(k8sClient client.Client) error {
	logf.Log.Info("Deploying kuadrant operator", "version", KUADRANT_OPERATOR_VERSION)

	//apiVersion: operators.coreos.com/v1alpha1
	//kind: Subscription
	//metadata:
	//  name: my-kuadrant-operator
	//  namespace: operators
	//spec:
	//  channel: stable
	//  name: kuadrant-operator
	//  source: operatorhubio-catalog
	//  sourceNamespace: olm
	//
	subs := &operators.Subscription{
		TypeMeta:   metav1.TypeMeta{APIVersion: "operators.coreos.com/v1alpha1", Kind: "Subscription"},
		ObjectMeta: metav1.ObjectMeta{Name: "kuadrant-operator", Namespace: "operators"},
		Spec: &operators.SubscriptionSpec{
			Channel:                "stable",
			Package:                "kuadrant-operator",
			CatalogSource:          "operatorhubio-catalog",
			CatalogSourceNamespace: "olm",
			StartingCSV:            fmt.Sprintf("kuadrant-operator.%s", KUADRANT_OPERATOR_VERSION),
		},
	}

	err := utils.CreateOnlyK8SObject(k8sClient, subs)
	if err != nil {
		return err
	}

	return nil
}

func deployKuadrant(k8sClient client.Client) error {
	kuadrant := &kuadrantoperator.Kuadrant{
		TypeMeta:   metav1.TypeMeta{APIVersion: "kuadrant.io/v1beta1", Kind: "Kuadrant"},
		ObjectMeta: metav1.ObjectMeta{Name: "kuadrant", Namespace: installNamespace},
		Spec:       kuadrantoperator.KuadrantSpec{},
	}

	err := utils.CreateOnlyK8SObject(k8sClient, kuadrant)
	if err != nil {
		return err
	}

	return nil
}

func createNamespace(k8sClient client.Client) error {
	nsObj := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: installNamespace},
	}
	logf.Log.Info("Creating kuadrant namespace", "name", installNamespace)
	err := utils.CreateOnlyK8SObject(k8sClient, nsObj)
	if err != nil {
		return err
	}

	retryInterval := time.Second * 2
	timeout := time.Second * 20
	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		err := k8sClient.Get(context.Background(), types.NamespacedName{Name: installNamespace}, &corev1.Namespace{})
		if err != nil && apierrors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	})
}
