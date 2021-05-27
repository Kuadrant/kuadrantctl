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
	"fmt"
	"time"

	networkingv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"github.com/spf13/cobra"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioSecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kuadrant/kuadrantctl/authorinomanifests"
	"github.com/kuadrant/kuadrantctl/istiomanifests"
	"github.com/kuadrant/kuadrantctl/kuadrantmanifests"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	installKubeConfig string
	// TODO(eastizle): namespace from command line param
	installNamespace string = "kuadrant-system"
	// TODO(eastizle): kuadrant controller image from command option
	installControllerImage string = "quay.io/eastizle/kuadrant-controller:v0.0.1-pre2"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Applies an kuadrant manifest, installing or reconfiguring kuadrant on a cluster",
	Long:  "The install command generates an Istio install manifest and applies it to a cluster.",
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

func setupScheme() error {
	err := apiextensionsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = apiextensionsv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = istio.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = istioSecurity.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = networkingv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	return nil
}

func installRun(cmd *cobra.Command, args []string) error {
	err := setupScheme()
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

	err = deployKuadrant(k8sClient)
	if err != nil {
		return err
	}

	err = waitForDeployments(k8sClient)
	if err != nil {
		return err
	}

	logf.Log.Info("kuadrant successfully deployed", "version", installControllerImage)

	return nil
}

func createOrUpdate(k8sClient client.Client) utils.DecodeCallback {
	return func(obj runtime.Object) error {
		return utils.CreateOrUpdateK8SObject(k8sClient, obj)
	}
}

func createOnly(k8sClient client.Client) utils.DecodeCallback {
	return func(obj runtime.Object) error {
		return utils.CreateOnlyK8SObject(k8sClient, obj)
	}
}

func waitForDeployments(k8sClient client.Client) error {
	retryInterval := time.Second * 5
	timeout := time.Minute * 2
	expectedDeployments := 3
	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		ready, err := utils.CheckForDeploymentsReady(installNamespace, k8sClient, expectedDeployments)
		if err != nil {
			return false, err
		}
		if !ready {
			logf.Log.Info("Waiting for deployments full availability")
		}
		return ready, nil
	})
}

func deployKuadrant(k8sClient client.Client) error {
	data, err := kuadrantmanifests.Content()
	if err != nil {
		return err
	}

	err = utils.DecodeFile(data, scheme.Scheme, createOnly(k8sClient))
	if err != nil {
		return err
	}

	// Update deployment
	managerDeployment := &appsv1.Deployment{}
	retryInterval := time.Second * 2
	timeout := time.Second * 20
	err = wait.Poll(retryInterval, timeout, func() (bool, error) {
		err := k8sClient.Get(context.Background(),
			types.NamespacedName{Namespace: installNamespace, Name: "kuadrant-controller-manager"},
			managerDeployment)
		if err != nil && apierrors.IsNotFound(err) {
			return false, nil
		}
		return true, err
	})
	if err != nil {
		return err
	}

	// Update image
	patchStr := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name": "manager","image":"%s","imagePullPolicy":"IfNotPresent"}]}}}}`, installControllerImage)
	patch := []byte(patchStr)
	err = k8sClient.Patch(context.Background(),
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: managerDeployment.GetNamespace(), Name: managerDeployment.GetName()}},
		client.RawPatch(types.StrategicMergePatchType, patch))
	logf.Log.Info("patch kuadrant controller deployment", "name", managerDeployment.GetName(), "image", installControllerImage, "error", err)
	if err != nil {
		return err
	}

	return nil
}

func deployAuthorizationProvider(k8sClient client.Client) error {
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

func init() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	// TODO(eastizle): add context flag to switch between kubeconfig contexts
	// It would require using config.GetConfigWithContext(context string) (*rest.Config, error)
	installCmd.PersistentFlags().StringVarP(&installKubeConfig, "kubeconfig", "", "", "Kubernetes configuration file")
	rootCmd.AddCommand(installCmd)
}
