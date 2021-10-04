/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"flag"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
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

var (
	installKubeConfig string
	// TODO(eastizle): namespace from command line param
	installNamespace string = "kuadrant-system"
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

	err = createNamespace(k8sClient)
	if err != nil {
		return err
	}

	err = deployIngressProvider(k8sClient)
	if err != nil {
		return err
	}

	err = deployAuthorizationProvider(k8sClient)
	if err != nil {
		return err
	}

	err = deployRateLimitProvider(k8sClient)
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

func createOnly(k8sClient client.Client) utils.DecodeCallback {
	return func(obj runtime.Object) error {
		return utils.CreateOnlyK8SObject(k8sClient, obj)
	}
}

func waitForDeployments(k8sClient client.Client) error {
	retryInterval := time.Second * 5
	timeout := time.Minute * 2

	deploymentKeys := []types.NamespacedName{
		types.NamespacedName{Name: "kuadrant-gateway", Namespace: installNamespace},
		types.NamespacedName{Name: "authorino-controller-manager", Namespace: installNamespace},
		types.NamespacedName{Name: "limitador", Namespace: installNamespace},
		types.NamespacedName{Name: "kuadrant-controller-manager", Namespace: installNamespace},
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

func deployKuadrant(k8sClient client.Client) error {
	kuadrantControllerVersion, err := utils.KuadrantControllerImage()
	if err != nil {
		return err
	}
	logf.Log.Info("Deploying kuadrant controller", "version", kuadrantControllerVersion)

	data, err := kuadrantmanifests.Content()
	if err != nil {
		return err
	}

	err = utils.DecodeFile(data, scheme.Scheme, createOnly(k8sClient))
	if err != nil {
		return err
	}

	return nil
}

func deployAuthorizationProvider(k8sClient client.Client) error {
	authorinoVersion, err := utils.AuthorinoImage()
	if err != nil {
		return err
	}
	logf.Log.Info("Deploying authorino", "version", authorinoVersion)

	data, err := authorinomanifests.Content()
	if err != nil {
		return err
	}

	err = utils.DecodeFile(data, scheme.Scheme, createOnly(k8sClient))
	if err != nil {
		return err
	}

	return nil
}

func deployIngressProvider(k8sClient client.Client) error {
	istioVersion, err := utils.IstioImage()
	if err != nil {
		return err
	}
	logf.Log.Info("Deploying istio", "version", istioVersion)
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
		err = utils.DecodeFile(data, scheme.Scheme, createOnly(k8sClient))
		if err != nil {
			return err
		}
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

func deployRateLimitProvider(k8sClient client.Client) error {
	limitadorOperatorVersion, err := utils.LimitadorOperatorImage()
	if err != nil {
		return err
	}
	logf.Log.Info("Deploying limitador operator", "version", limitadorOperatorVersion)

	data, err := limitadormanifests.OperatorContent()
	if err != nil {
		return err
	}
	err = utils.DecodeFile(data, scheme.Scheme, createOnly(k8sClient))
	if err != nil {
		return err
	}

	limitadorObj := limitador.Limitador(installNamespace)
	logf.Log.Info("Deploying limitador instance", "version", *limitadorObj.Spec.Version)
	return utils.CreateOnlyK8SObject(k8sClient, limitadorObj)
}
