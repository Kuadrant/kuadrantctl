package cmd

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kuadrant/kuadrantctl/kuadrantmanifests"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

var (
	testK8sClient client.Client
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var _ = BeforeSuite(func() {
	By("Before suite")

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	configuration, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	err = apiextensionsv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	testK8sClient, err = client.New(configuration, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(testK8sClient).NotTo(BeNil())

	// install CRDs
	data, err := kuadrantmanifests.Content()
	Expect(err).NotTo(HaveOccurred())

	createOnlyCRDs := func(obj runtime.Object) error {
		k8sObj, ok := obj.(client.Object)
		if !ok {
			return errors.New("runtime.Object could not be casted to client.Object")
		}

		if obj.GetObjectKind().GroupVersionKind().Group != apiextensionsv1beta1.GroupName &&
			obj.GetObjectKind().GroupVersionKind().Group != apiextensionsv1.GroupName {
			// only CRD's
			return nil
		}

		err := testK8sClient.Create(context.Background(), k8sObj)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
		return nil
	}

	err = utils.DecodeFile(data, scheme.Scheme, createOnlyCRDs)
	Expect(err).NotTo(HaveOccurred())
})
