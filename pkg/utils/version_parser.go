package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kuadrant/kuadrantctl/authorinomanifests"
	"github.com/kuadrant/kuadrantctl/istiomanifests"
	"github.com/kuadrant/kuadrantctl/kuadrantmanifests"
	"github.com/kuadrant/kuadrantctl/limitadormanifests"
)

func IstioImage() (string, error) {
	istioImage := "unknown"

	istioParser := func(obj runtime.Object) error {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.GetName() == "istiod" {
				istioImage = deployment.Spec.Template.Spec.Containers[0].Image
			}
		}
		return nil
	}

	istioPilotContent, err := istiomanifests.PilotContent()
	if err != nil {
		return "", err
	}

	err = DecodeFile(istioPilotContent, scheme.Scheme, istioParser)
	if err != nil {
		return "", err
	}

	return istioImage, nil
}

func AuthorinoImage() (string, error) {
	image := "unknown"

	authorinoParser := func(obj runtime.Object) error {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.GetName() == "authorino-controller-manager" {
				for _, container := range deployment.Spec.Template.Spec.Containers {
					if container.Name == "manager" {
						image = container.Image
					}
				}
			}
		}
		return nil
	}

	content, err := authorinomanifests.Content()
	if err != nil {
		return "", err
	}

	err = DecodeFile(content, scheme.Scheme, authorinoParser)
	if err != nil {
		return "", err
	}

	return image, nil
}

func LimitadorOperatorImage() (string, error) {
	image := "unknown"

	parser := func(obj runtime.Object) error {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.GetName() == "limitador-operator-controller-manager" {
				for _, container := range deployment.Spec.Template.Spec.Containers {
					if container.Name == "manager" {
						image = container.Image
					}
				}
			}
		}
		return nil
	}

	content, err := limitadormanifests.OperatorContent()
	if err != nil {
		return "", err
	}

	err = DecodeFile(content, scheme.Scheme, parser)
	if err != nil {
		return "", err
	}

	return image, nil
}

func KuadrantControllerImage() (string, error) {
	image := "unknown"

	parser := func(obj runtime.Object) error {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.GetName() == "kuadrant-controller-manager" {
				for _, container := range deployment.Spec.Template.Spec.Containers {
					if container.Name == "manager" {
						image = container.Image
					}
				}
			}
		}
		return nil
	}

	content, err := kuadrantmanifests.Content()
	if err != nil {
		return "", err
	}

	err = DecodeFile(content, scheme.Scheme, parser)
	if err != nil {
		return "", err
	}

	return image, nil
}
