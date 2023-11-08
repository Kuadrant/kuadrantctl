package utils

import (
	kuadrantoperator "github.com/kuadrant/kuadrant-operator/api/v1beta1"
	operators "github.com/operator-framework/api/pkg/operators/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func SetupScheme() error {
	err := apiextensionsv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = operators.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	err = kuadrantoperator.AddToScheme(scheme.Scheme)
	if err != nil {
		return err
	}

	return nil
}
