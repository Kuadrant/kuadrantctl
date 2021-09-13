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
package utils

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateOrUpdateK8SObject(k8sClient client.Client, obj runtime.Object) error {
	k8sObj, ok := obj.(client.Object)
	if !ok {
		return errors.New("runtime.Object could not be casted to client.Object")
	}

	err := k8sClient.Create(context.Background(), k8sObj)
	logf.Log.V(1).Info("create resource", "GKV", k8sObj.GetObjectKind().GroupVersionKind(), "name", k8sObj.GetName(), "error", err)
	if err == nil {
		return nil
	}

	if !apierrors.IsAlreadyExists(err) {
		return err
	}

	// Already exists
	currentObj := k8sObj.DeepCopyObject()
	k8sCurrentObj, ok := currentObj.(client.Object)
	if !ok {
		return errors.New("runtime.Object could not be casted to client.Object")
	}
	err = k8sClient.Get(context.Background(), client.ObjectKeyFromObject(k8sObj), k8sCurrentObj)
	if err != nil {
		return err
	}

	objCopy := k8sObj.DeepCopyObject()

	objCopyMetadata, err := meta.Accessor(objCopy)
	if err != nil {
		return err
	}

	objCopyMetadata.SetResourceVersion(k8sCurrentObj.GetResourceVersion())

	k8sObjCopy, ok := objCopy.(client.Object)
	if !ok {
		return errors.New("runtime.Object could not be casted to client.Object")
	}

	err = k8sClient.Update(context.Background(), k8sObjCopy)
	logf.Log.Info("update resource", "GKV", k8sObj.GetObjectKind().GroupVersionKind(), "name", k8sObj.GetName(), "error", err)
	return err
}

func CreateOnlyK8SObject(k8sClient client.Client, obj runtime.Object) error {
	k8sObj, ok := obj.(client.Object)
	if !ok {
		return errors.New("runtime.Object could not be casted to client.Object")
	}
	k8sObjKind := k8sObj.DeepCopyObject().GetObjectKind()

	err := k8sClient.Create(context.Background(), k8sObj)
	logf.Log.V(1).Info("create resource", "GKV", k8sObjKind.GroupVersionKind(), "name", k8sObj.GetName(), "error", err)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Omit error
			logf.Log.Info("Already exists", "GKV", k8sObjKind.GroupVersionKind(), "name", k8sObj.GetName())
		} else {
			return err
		}
	}
	return nil
}

func DeleteK8SObject(k8sClient client.Client, obj runtime.Object) error {
	k8sObj, ok := obj.(client.Object)
	if !ok {
		return errors.New("runtime.Object could not be casted to client.Object")
	}
	k8sObjKind := k8sObj.DeepCopyObject().GetObjectKind()

	err := k8sClient.Delete(context.Background(), k8sObj)
	logf.Log.V(1).Info("delete resource", "GKV", k8sObjKind.GroupVersionKind(), "name", k8sObj.GetName(), "error", err)
	if err != nil && !apierrors.IsNotFound(err) {
		// Omit NotFound error
		return err
	}
	return nil
}

// IsDeploymentAvailable returns true when the provided Deployment
// has the "Available" condition set to true
func IsDeploymentAvailable(dc *appsv1.Deployment) bool {
	dcConditions := dc.Status.Conditions
	for _, dcCondition := range dcConditions {
		if dcCondition.Type == appsv1.DeploymentAvailable && dcCondition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func CheckDeploymentAvailable(k8sClient client.Client, key types.NamespacedName) (bool, error) {
	existingDeployment := &appsv1.Deployment{}
	err := k8sClient.Get(context.Background(), key, existingDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logf.Log.Info("Deployment not available", "name", key.Name)
			return false, nil
		}

		return false, err
	}

	if !IsDeploymentAvailable(existingDeployment) {
		logf.Log.Info("Waiting for full availability", "Deployment", existingDeployment.GetName(),
			"available replicas", existingDeployment.Status.AvailableReplicas, "desired replics",
			*existingDeployment.Spec.Replicas)
		return false, nil
	}

	logf.Log.Info("Deployment available", "name", existingDeployment.GetName())
	return true, nil
}
