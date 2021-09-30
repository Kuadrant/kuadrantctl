package utils

import (
	"fmt"
	"reflect"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func(existing, desired client.Object) (bool, error)

func CreateOnlyMutator(existing, desired client.Object) (bool, error) {
	return false, nil
}

func KuadrantAPIBasicMutator(existingObj, desiredObj client.Object) (bool, error) {
	existing, ok := existingObj.(*kctlrv1beta1.API)
	if !ok {
		return false, fmt.Errorf("%T is not a *kctlrv1beta1.API", existingObj)
	}
	desired, ok := desiredObj.(*kctlrv1beta1.API)
	if !ok {
		return false, fmt.Errorf("%T is not a *kctlrv1beta1.API", desiredObj)
	}

	updated := false
	if !reflect.DeepEqual(existing.Spec, desired.Spec) {
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
