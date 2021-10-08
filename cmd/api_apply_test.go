package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
	"sigs.k8s.io/yaml"
)

var _ = Describe("the api apply command", func() {
	var testNamespace string

	BeforeEach(func() {
		var generatedTestNamespace = "test-namespace-" + uuid.New().String()

		namespace := &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
			ObjectMeta: metav1.ObjectMeta{Name: generatedTestNamespace},
		}

		// Add any setup steps that needs to be executed before each test
		err := testK8sClient.Create(context.Background(), namespace)
		Expect(err).ToNot(HaveOccurred())

		existingNamespace := &corev1.Namespace{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: generatedTestNamespace}, existingNamespace)
			return err == nil
		}, 5*time.Minute, 5*time.Second).Should(BeTrue())

		testNamespace = existingNamespace.Name
	})

	AfterEach(func() {
		desiredTestNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}
		// Add any teardown steps that needs to be executed after each test
		err := testK8sClient.Delete(context.Background(), desiredTestNamespace, client.PropagationPolicy(metav1.DeletePropagationForeground))

		Expect(err).ToNot(HaveOccurred())

		existingNamespace := &corev1.Namespace{}
		Eventually(func() bool {
			err := testK8sClient.Get(context.Background(), types.NamespacedName{Name: testNamespace}, existingNamespace)
			if err != nil && apierrors.IsNotFound(err) {
				return true
			}
			return false
		}, 5*time.Minute, 5*time.Second).Should(BeTrue())

	})

	Context("when the match type is invalid", func() {
		It("should return not found error", func() {
			cmdLine := fmt.Sprintf("api apply --match-path /v1 --match-path-type unknown --service-name someservice -n %s",
				testNamespace)
			rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not valid match-path-type"))
		})
	})

	Context("when the service is not found", func() {
		It("should return not found error", func() {
			cmdLine := fmt.Sprintf("api apply --service-name someservice -n %s --oas testdata/invalid_oas.yaml",
				testNamespace)
			rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("services \"someservice\" not found"))
		})
	})

	Context("when the service exists", func() {
		var serviceName string
		BeforeEach(func() {
			var generatedServiceName = "test-service-" + uuid.New().String()

			service := &corev1.Service{
				TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
				ObjectMeta: metav1.ObjectMeta{Name: generatedServiceName, Namespace: testNamespace},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{"key": "value"},
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			}

			// Add any setup steps that needs to be executed before each test
			err := testK8sClient.Create(context.Background(), service)
			Expect(err).ToNot(HaveOccurred())

			serviceName = service.Name
		})

		Context("when the OAS is invalid", func() {
			It("should return validation error", func() {
				cmdLine := fmt.Sprintf("api apply --service-name %s -n %s --oas testdata/invalid_oas.yaml",
					serviceName, testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("OpenAPI validation error"))
			})
		})

		Context("oas is given", func() {
			It("API should contain the oas", func() {
				cmdLine := fmt.Sprintf("api apply --service-name %s -n %s --oas testdata/petstore.yaml",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())

				expectedAPIName := serviceName
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(existingAPI.Spec.Mappings.OAS).NotTo(BeNil())
			})
		})

		Context("when tag is given", func() {
			It("API name should contain the tag", func() {
				cmdLine := fmt.Sprintf("api apply --tag production --service-name %s -n %s --oas testdata/petstore.yaml",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())

				expectedAPIName := serviceName + "." + "production"
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when port number is given", func() {
			It("API destination port has the port", func() {
				cmdLine := fmt.Sprintf("api apply --port 78 --service-name %s -n %s --oas testdata/petstore.yaml",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())
				expectedAPIName := serviceName
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(*existingAPI.Spec.Destination.ServiceReference.Port).Should(Equal(int32(78)))
			})
		})

		Context("when port name is given", func() {
			It("API destination port has the port", func() {
				cmdLine := fmt.Sprintf("api apply --port http --service-name %s -n %s --oas testdata/petstore.yaml",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())
				expectedAPIName := serviceName
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(*existingAPI.Spec.Destination.ServiceReference.Port).Should(Equal(int32(80)))
			})
		})

		Context("when match path is given", func() {
			It("API mapping has match path", func() {
				cmdLine := fmt.Sprintf("api apply --service-name %s -n %s --match-path /v1 --match-path-type Exact",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))
				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())
				expectedAPIName := serviceName
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).ToNot(HaveOccurred())
				Expect(existingAPI.Spec.Mappings.HTTPPathMatch).NotTo(BeNil())
				Expect(existingAPI.Spec.Mappings.HTTPPathMatch.Value).NotTo(BeNil())
				Expect(*existingAPI.Spec.Mappings.HTTPPathMatch.Value).Should(Equal("/v1"))
				Expect(existingAPI.Spec.Mappings.HTTPPathMatch.Type).NotTo(BeNil())
				Expect(*existingAPI.Spec.Mappings.HTTPPathMatch.Type).Should(Equal(gatewayapiv1alpha1.PathMatchType("Exact")))
			})
		})

		Context("with the --to-stdout flg", func() {
			It("should write to stdout", func() {
				cmdLine := fmt.Sprintf("api apply --to-stdout --service-name %s -n %s --match-path /v1",
					serviceName,
					testNamespace)
				rootCmd := GetRootCmd(strings.Split(cmdLine, " "))

				var out bytes.Buffer
				rootCmd.SetOut(&out)
				rootCmd.SetErr(&out)

				err := rootCmd.Execute()
				Expect(err).ToNot(HaveOccurred())

				expectedAPIName := serviceName
				existingAPI := &kctlrv1beta1.API{}
				err = testK8sClient.Get(context.Background(),
					types.NamespacedName{Name: expectedAPIName, Namespace: testNamespace},
					existingAPI)
				Expect(err).To(HaveOccurred())
				Expect(apierrors.IsNotFound(err)).To(BeTrue())

				api := &kctlrv1beta1.API{}
				err = yaml.Unmarshal(out.Bytes(), api)
				Expect(err).ToNot(HaveOccurred())
				Expect(api.TypeMeta.Kind).Should(Equal("API"))
				Expect(api.TypeMeta.APIVersion).Should(Equal(kctlrv1beta1.GroupVersion.String()))
			})
		})
	})
})
