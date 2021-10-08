//Copyright 2021 Red Hat, Inc.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayapiv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	kubeConfig               string
	apiApplyServiceNamespace string
	apiApplyServiceName      string
	apiApplyScheme           string
	apiApplyAPIName          string
	apiApplyTag              string
	apiApplyPortStr          string
	apiApplyMatchPath        string
	apiApplyMatchPathTypeStr string
	apiApplyMatchPathType    gatewayapiv1alpha1.PathMatchType
	apiApplyOAS              string
	apiApplyToStdout         bool
)

func apiApplyCommand() *cobra.Command {
	apiApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Applies a Kuadrant API, installing on a cluster",
		Long:  "The apply command allows easily to create and update existing *kuadrant API* custom resources",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Required to have controller-runtim config package read the kubeconfig arg
			err := flag.CommandLine.Parse([]string{"-kubeconfig", kubeConfig})
			if err != nil {
				return err
			}

			if apiApplyScheme != "http" && apiApplyScheme != "https" {
				return errors.New("not valid scheme. Only ['http', 'https'] allowed")
			}


			pathMatchType := gatewayapiv1alpha1.PathMatchType(apiApplyMatchPathTypeStr)
			switch pathMatchType {
			case gatewayapiv1alpha1.PathMatchExact, gatewayapiv1alpha1.PathMatchPrefix, gatewayapiv1alpha1.PathMatchRegularExpression:
				apiApplyMatchPathType = pathMatchType
			default:
				return fmt.Errorf("not valid match-path-type. Only ['%s', '%s', '%s'] allowed",
					gatewayapiv1alpha1.PathMatchExact, gatewayapiv1alpha1.PathMatchPrefix,
					gatewayapiv1alpha1.PathMatchRegularExpression)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApiApply(cmd, args)
		},
	}

	// TODO(eastizle): add context flag to switch between kubeconfig contexts
	// It would require using config.GetConfigWithContext(context string) (*rest.Config, error)
	apiApplyCmd.PersistentFlags().StringVar(&kubeConfig, "kubeconfig", "", "Kubernetes configuration file")
	apiApplyCmd.Flags().StringVar(&apiApplyServiceName, "service-name", "", "Service name (required)")
	err := apiApplyCmd.MarkFlagRequired("service-name")
	if err != nil {
		panic(err)
	}
	apiApplyCmd.Flags().StringVarP(&apiApplyServiceNamespace, "namespace", "n", "", "Service namespace (required)")
	err = apiApplyCmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}
	apiApplyCmd.Flags().StringVar(&apiApplyScheme, "scheme", "http", "Either HTTP or HTTPS specifies how the kuadrant gateway will connect to this API")
	apiApplyCmd.Flags().StringVar(&apiApplyAPIName, "api-name", "", "If not set, the name of the API can be matched with the service name")
	apiApplyCmd.Flags().StringVar(&apiApplyTag, "tag", "", "A special tag used to distinguish this deployment between several instances of the API")
	apiApplyCmd.Flags().StringVar(&apiApplyPortStr, "port", "", "Only required if there are multiple ports in the service. Either the Name of the port or the Number")
	apiApplyCmd.Flags().StringVar(&apiApplyMatchPath, "match-path", "/", "Define a single specific path, prefix or regex")
	apiApplyCmd.Flags().StringVar(&apiApplyMatchPathTypeStr, "match-path-type", "Prefix", "Specifies how to match against the matchpath value. Accepted values are Exact, Prefix and RegularExpression. Defaults to Prefix")
	apiApplyCmd.Flags().StringVar(&apiApplyOAS, "oas", "", "/path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR -")
	apiApplyCmd.Flags().BoolVar(&apiApplyToStdout, "to-stdout", false, "Serialize the kuadrant API object in stdout instead of applying to the cluster")

	return apiApplyCmd
}

func runApiApply(cmd *cobra.Command, args []string) error {
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

	serviceKey := client.ObjectKey{Name: apiApplyServiceName, Namespace: apiApplyServiceNamespace}

	service := &corev1.Service{}
	if err := k8sClient.Get(context.TODO(), serviceKey, service); err != nil {
		// the service must exist
		return err
	}

	apiName := service.GetName()
	if apiApplyAPIName != "" {
		apiName = apiApplyAPIName
	}

	if apiApplyTag != "" {
		apiName = fmt.Sprintf("%s.%s", apiName, apiApplyTag)
	}

	var destinationPort int32
	if apiApplyPortStr != "" {
		// check if the port is a number already.
		if num, err := strconv.ParseInt(apiApplyPortStr, 10, 32); err == nil {
			destinationPort = int32(num)
		} else {
			// As the port is name, resolv the port from the service
			for _, p := range service.Spec.Ports {
				if p.Name == apiApplyPortStr {
					destinationPort = p.Port
					break
				}
			}
		}
	} else {
		// let's check if the service has only one port, if that's the case,
		// default to it.
		if len(service.Spec.Ports) != 1 {
			return errors.New("multple ports found in the service. Use --port to specify")
		}

		destinationPort = service.Spec.Ports[0].Port
	}

	// If we reach this point and the Port is still nil, this means bad news
	if destinationPort == 0 {
		return errors.New("port is missing or invalid")
	}

	api := &kctlrv1beta1.API{
		TypeMeta: metav1.TypeMeta{
			Kind:       kctlrv1beta1.APIKind,
			APIVersion: kctlrv1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			Namespace: apiApplyServiceNamespace,
		},
		Spec: kctlrv1beta1.APISpec{
			Destination: kctlrv1beta1.Destination{
				Schema: apiApplyScheme,
				ServiceReference: apiextensionsv1.ServiceReference{
					Namespace: service.GetNamespace(),
					Name:      service.GetName(),
					Port:      &destinationPort,
				},
			},
			// Default value for mappings is path match / with Prefix type
			Mappings: kctlrv1beta1.APIMappings{
				HTTPPathMatch: &gatewayapiv1alpha1.HTTPPathMatch{
					Type: &apiApplyMatchPathType, Value: &apiApplyMatchPath,
				},
			},
		},
	}

	if apiApplyOAS != "" {
		dataRaw, err := utils.ReadExternalResource(apiApplyOAS)
		if err != nil {
			return err
		}

		err = utils.ValidateOAS3(dataRaw)
		if err != nil {
			return err
		}

		dataTmp := string(dataRaw)
		api.Spec.Mappings = kctlrv1beta1.APIMappings{OAS: &dataTmp}
	}

	// Add owner reference. This is not a controller owner reference
	err = controllerutil.SetOwnerReference(service, api, scheme.Scheme)
	if err != nil {
		return err
	}

	if apiApplyToStdout {
		// In short, this library first converts YAML to JSON using go-yaml
		// and then uses json.Marshal and json.Unmarshal to convert to or from the struct
		yamlData, err := yaml.Marshal(api)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(yamlData))
	} else {
		err = utils.ReconcileKuadrantAPI(k8sClient, api, utils.KuadrantAPIBasicMutator)
		if err != nil {
			return err
		}
	}

	return nil
}
