package utils

import (
	"context"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ReconcileResource attempts to mutate the existing state
// in order to match the desired state. The object's desired state must be reconciled
// with the existing state inside the passed in callback MutateFn.
//
// obj: Object of the same type as the 'desired' object.
//            Used to read the resource from the kubernetes cluster.
//            Could be zero-valued initialized object.
// desired: Object representing the desired state
//
// It returns an error.
func ReconcileResource(k8sClient client.Client, ctx context.Context, obj, desired client.Object, mutateFn MutateFn) error {
	key := client.ObjectKeyFromObject(desired)

	if err := k8sClient.Get(ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		// Not found

		// k8s client could remove TypeMeta after object create
		k8sObjKind := desired.DeepCopyObject().GetObjectKind()
		err = k8sClient.Create(ctx, desired)
		logf.Log.Info("create API object", "GKV", k8sObjKind.GroupVersionKind(), "name", desired.GetName(), "error", err)
		return err
	}

	update, err := mutateFn(obj, desired)
	if err != nil {
		return err
	}

	if update {
		err := k8sClient.Update(ctx, obj)
		logf.Log.Info("update API object", "GKV", desired.GetObjectKind().GroupVersionKind(), "name", obj.GetName(), "error", err)
		return err
	}

	logf.Log.Info("API object is up to date. Nothing to do", "GKV", desired.GetObjectKind().GroupVersionKind(), "name", obj.GetName())

	return nil
}

func ReconcileKuadrantAPI(k8sClient client.Client, desired *kctlrv1beta1.API, mutatefn MutateFn) error {
	return ReconcileResource(k8sClient, context.TODO(), &kctlrv1beta1.API{}, desired, mutatefn)
}
