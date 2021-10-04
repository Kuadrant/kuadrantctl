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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("the install command", func() {
	BeforeEach(func() {
		namespace := &corev1.Namespace{}
		Expect(testK8sClient).NotTo(BeNil())
		err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: installNamespace}, namespace)
		Expect(err).To(HaveOccurred())
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	AfterEach(func() {
		desiredTestNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: installNamespace}}
		err := testK8sClient.Delete(context.Background(), desiredTestNamespace, client.PropagationPolicy(metav1.DeletePropagationForeground))
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() bool {
			existingNamespace := &corev1.Namespace{}
			err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: installNamespace}, existingNamespace)
			if err != nil && apierrors.IsNotFound(err) {
				return true
			}
			return false
		}, 2*time.Minute, 5*time.Second).Should(BeTrue())

	})

	It("should deploy kuadrant dependency services", func() {
		rootCmd := GetRootCmd([]string{"install"})
		err := rootCmd.Execute()
		Expect(err).ToNot(HaveOccurred())
	})
})
