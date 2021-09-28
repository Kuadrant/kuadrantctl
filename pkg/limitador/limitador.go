package limitador

import (
	limitadorv1alpha1 "github.com/kuadrant/limitador-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Limitador(ns string) *limitadorv1alpha1.Limitador {
	tmpVersion := "0.4.0"
	return &limitadorv1alpha1.Limitador{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Limitador",
			APIVersion: "limitador.kuadrant.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "limitador",
			Namespace: ns,
		},
		Spec: limitadorv1alpha1.LimitadorSpec{
			Version: &tmpVersion,
		},
	}
}
