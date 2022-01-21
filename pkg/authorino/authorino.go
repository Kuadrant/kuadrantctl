package authorino

import (
	authorinov1beta1 "github.com/kuadrant/authorino-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Authorino(ns string) *authorinov1beta1.Authorino {
	tlsEnabledTmp := false
	return &authorinov1beta1.Authorino{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Authorino",
			APIVersion: "operator.authorino.kuadrant.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "authorino",
			Namespace: ns,
		},
		Spec: authorinov1beta1.AuthorinoSpec{
			Image:       "quay.io/3scale/authorino:v0.7.0",
			ClusterWide: true,
			Listener: authorinov1beta1.Listener{
				Tls: authorinov1beta1.Tls{
					Enabled: &tlsEnabledTmp,
				},
			},
			OIDCServer: authorinov1beta1.OIDCServer{
				Tls: authorinov1beta1.Tls{
					Enabled: &tlsEnabledTmp,
				},
			},
			SecretLabelSelectors: "authorino.kuadrant.io/managed-by=authorino",
		},
	}
}
