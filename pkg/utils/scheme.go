package utils

import (
	authorinov1beta1 "github.com/kuadrant/authorino-operator/api/v1beta1"
	networkingv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	limitadorv1alpha1 "github.com/kuadrant/limitador-operator/api/v1alpha1"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func SetupScheme() error {
	err := apiextensionsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = istio.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = networkingv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = limitadorv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = authorinov1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	return nil
}
