package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	supabasev1alpha1 "github.com/strrl/supabase-operator/api/v1alpha1"
)

var _ = Describe("SupabaseProject Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		supabaseproject := &supabasev1alpha1.SupabaseProject{}

		BeforeEach(func() {
			By("creating the required database secret")
			dbSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-secret",
					Namespace: "default",
				},
				StringData: map[string]string{
					"host":     "postgres.example.com",
					"port":     "5432",
					"database": "supabase",
					"username": "postgres",
					"password": "test-password",
				},
			}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-db-secret", Namespace: "default"}, &corev1.Secret{})
			if errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, dbSecret)).To(Succeed())
			}

			By("creating the required storage secret")
			storageSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-storage-secret",
					Namespace: "default",
				},
				StringData: map[string]string{
					"endpoint":   "s3.amazonaws.com",
					"region":     "us-east-1",
					"bucket":     "test-bucket",
					"access-key": "test-access-key",
					"secret-key": "test-secret-key",
				},
			}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-storage-secret", Namespace: "default"}, &corev1.Secret{})
			if errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, storageSecret)).To(Succeed())
			}

			By("creating the custom resource for the Kind SupabaseProject")
			err = k8sClient.Get(ctx, typeNamespacedName, supabaseproject)
			if err != nil && errors.IsNotFound(err) {
				resource := &supabasev1alpha1.SupabaseProject{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: supabasev1alpha1.SupabaseProjectSpec{
						ProjectID: "test-project",
						Database: supabasev1alpha1.DatabaseConfig{
							SecretRef: corev1.SecretReference{
								Name: "test-db-secret",
							},
						},
						Storage: supabasev1alpha1.StorageConfig{
							SecretRef: corev1.SecretReference{
								Name: "test-storage-secret",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &supabasev1alpha1.SupabaseProject{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance SupabaseProject")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the database secret")
			dbSecret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-db-secret", Namespace: "default"}, dbSecret)
			if err == nil {
				Expect(k8sClient.Delete(ctx, dbSecret)).To(Succeed())
			}

			By("Cleanup the storage secret")
			storageSecret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-storage-secret", Namespace: "default"}, storageSecret)
			if err == nil {
				Expect(k8sClient.Delete(ctx, storageSecret)).To(Succeed())
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &SupabaseProjectReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(100),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
