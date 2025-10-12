//go:build e2e
// +build e2e

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/strrl/supabase-operator/test/utils"
)

// namespace where the project is deployed in
const namespace = "supabase-operator-system"

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, installing CRDs,
	// and deploying the controller.
	BeforeAll(func() {
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to create namespace")

		By("labeling the namespace to enforce the restricted security policy")
		cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
			"pod-security.kubernetes.io/enforce=restricted")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to label namespace with restricted policy")

		By("installing CRDs")
		cmd = exec.Command("make", "install")
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to install CRDs")

		By("deploying the controller-manager")
		cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage))
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to deploy the controller-manager")
	})

	// After all tests have been executed, clean up by undeploying the controller, uninstalling CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("undeploying the controller-manager")
		cmd := exec.Command("make", "undeploy")
		_, _ = utils.Run(cmd)

		By("uninstalling CRDs")
		cmd = exec.Command("make", "uninstall")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			By("Fetching controller manager pod logs")
			cmd := exec.Command("kubectl", "logs", controllerPodName, "-n", namespace)
			controllerLogs, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Controller logs:\n %s", controllerLogs)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Controller logs: %s", err)
			}

			By("Fetching Kubernetes events")
			cmd = exec.Command("kubectl", "get", "events", "-n", namespace, "--sort-by=.lastTimestamp")
			eventsOutput, err := utils.Run(cmd)
			if err == nil {
				_, _ = fmt.Fprintf(GinkgoWriter, "Kubernetes events:\n%s", eventsOutput)
			} else {
				_, _ = fmt.Fprintf(GinkgoWriter, "Failed to get Kubernetes events: %s", err)
			}

			By("Fetching controller manager pod description")
			cmd = exec.Command("kubectl", "describe", "pod", controllerPodName, "-n", namespace)
			podDescription, err := utils.Run(cmd)
			if err == nil {
				fmt.Println("Pod description:\n", podDescription)
			} else {
				fmt.Println("Failed to describe controller pod")
			}
		}
	})

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	Context("Manager", func() {
		It("should run successfully", func() {
			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func(g Gomega) {
				// Get the name of the controller-manager pod
				cmd := exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Failed to retrieve controller-manager pod information")
				podNames := utils.GetNonEmptyLines(podOutput)
				g.Expect(podNames).To(HaveLen(1), "expected 1 controller pod running")
				controllerPodName = podNames[0]
				g.Expect(controllerPodName).To(ContainSubstring("controller-manager"))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")
			}
			Eventually(verifyControllerUp).Should(Succeed())
		})

		// +kubebuilder:scaffold:e2e-webhooks-checks
	})

	Context("SupabaseProject", func() {
		const testNamespace = "default"
		const projectName = "test-supabase-project"
		const dbSecretName = "test-db-secret"
		const storageSecretName = "test-storage-secret"

		BeforeEach(func() {
			By("creating database secret")
			createSecret(testNamespace, dbSecretName, map[string]string{
				"host":     "postgres.example.com",
				"port":     "5432",
				"database": "supabase",
				"username": "postgres",
				"password": "test-password",
			})

			By("creating storage secret")
			createSecret(testNamespace, storageSecretName, map[string]string{
				"endpoint":   "s3.amazonaws.com",
				"region":     "us-east-1",
				"bucket":     "test-bucket",
				"accessKeyId":     "test-access-key",
				"secretAccessKey": "test-secret-key",
			})
		})

		AfterEach(func() {
			By("cleaning up SupabaseProject")
			cmd := exec.Command("kubectl", "delete", "supabaseproject", projectName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)

			By("cleaning up secrets")
			cmd = exec.Command("kubectl", "delete", "secret", dbSecretName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)
			cmd = exec.Command("kubectl", "delete", "secret", storageSecretName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)
		})

		It("should successfully create and reconcile a SupabaseProject (T101)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("verifying the SupabaseProject is created")
			verifyResourceExists := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName, "-n", testNamespace)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}
			Eventually(verifyResourceExists).Should(Succeed())

			By("verifying JWT secret is created")
			verifyJWTSecret := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secret", fmt.Sprintf("%s-jwt", projectName), "-n", testNamespace)
				_, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
			}
			Eventually(verifyJWTSecret, 2*time.Minute).Should(Succeed())

			By("verifying status is updated to Running")
			verifyStatus := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName, "-n", testNamespace,
					"-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"))
			}
			Eventually(verifyStatus, 3*time.Minute).Should(Succeed())
		})

		It("should create all component deployments and services (T102)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			components := []string{"kong", "auth", "postgrest", "realtime", "storage", "meta"}
			for _, component := range components {
				componentName := component
				By(fmt.Sprintf("verifying %s deployment is created", componentName))
				verifyDeployment := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "deployment",
						fmt.Sprintf("%s-%s", projectName, componentName),
						"-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred())
				}
				Eventually(verifyDeployment, 2*time.Minute).Should(Succeed())

				By(fmt.Sprintf("verifying %s service is created", componentName))
				verifyService := func(g Gomega) {
					cmd := exec.Command("kubectl", "get", "service",
						fmt.Sprintf("%s-%s", projectName, componentName),
						"-n", testNamespace)
					_, err := utils.Run(cmd)
					g.Expect(err).NotTo(HaveOccurred())
				}
				Eventually(verifyService, 2*time.Minute).Should(Succeed())
			}
		})

		It("should generate JWT secrets with correct keys (T103)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("verifying JWT secret contains required keys")
			verifySecretKeys := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "secret",
					fmt.Sprintf("%s-jwt", projectName),
					"-n", testNamespace,
					"-o", "jsonpath={.data}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(ContainSubstring("jwt-secret"))
				g.Expect(output).To(ContainSubstring("anon-key"))
				g.Expect(output).To(ContainSubstring("service-role-key"))
			}
			Eventually(verifySecretKeys, 2*time.Minute).Should(Succeed())
		})

		It("should handle database initialization (T104)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("verifying reconciliation completes without database errors")
			verifyNoDBErrors := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName,
					"-n", testNamespace,
					"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).NotTo(Equal("False"))
			}
			Eventually(verifyNoDBErrors, 3*time.Minute).Should(Succeed())
		})

		It("should report failure when database secret is invalid (T105)", func() {
			By("creating invalid database secret")
			createSecret(testNamespace, "invalid-db-secret", map[string]string{
				"invalid-key": "invalid-value",
			})

			By("creating SupabaseProject with invalid secret")
			manifest := fmt.Sprintf(`
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: %s-invalid
  namespace: %s
spec:
  projectId: test-project-invalid
  database:
    secretRef:
      name: invalid-db-secret
  storage:
    secretRef:
      name: %s
`, projectName, testNamespace, storageSecretName)

			applyManifest(manifest)

			By("verifying phase transitions to Failed")
			verifyFailedStatus := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject",
					fmt.Sprintf("%s-invalid", projectName),
					"-n", testNamespace,
					"-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Failed"))
			}
			Eventually(verifyFailedStatus, 2*time.Minute).Should(Succeed())

			By("cleaning up invalid resources")
			cmd := exec.Command("kubectl", "delete", "supabaseproject",
				fmt.Sprintf("%s-invalid", projectName),
				"-n", testNamespace)
			_, _ = utils.Run(cmd)
			cmd = exec.Command("kubectl", "delete", "secret", "invalid-db-secret", "-n", testNamespace)
			_, _ = utils.Run(cmd)
		})

		It("should update component status correctly (T106a)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("verifying component status is populated")
			verifyComponentStatus := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName,
					"-n", testNamespace,
					"-o", "jsonpath={.status.components}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).NotTo(BeEmpty())
				g.Expect(output).To(ContainSubstring("kong"))
				g.Expect(output).To(ContainSubstring("auth"))
				g.Expect(output).To(ContainSubstring("postgrest"))
			}
			Eventually(verifyComponentStatus, 3*time.Minute).Should(Succeed())

			By("verifying Ready condition is True")
			verifyReadyCondition := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName,
					"-n", testNamespace,
					"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("True"))
			}
			Eventually(verifyReadyCondition, 3*time.Minute).Should(Succeed())
		})

		It("should handle spec updates (T106)", func() {
			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("waiting for initial reconciliation")
			time.Sleep(10 * time.Second)

			By("updating the SupabaseProject spec")
			patchJSON := `{"spec":{"kong":{"image":"kong:2.8.2"}}}`
			cmd := exec.Command("kubectl", "patch", "supabaseproject", projectName,
				"-n", testNamespace,
				"--type=merge",
				"-p", patchJSON)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("verifying observedGeneration is updated")
			verifyObservedGeneration := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName,
					"-n", testNamespace,
					"-o", "jsonpath={.status.observedGeneration}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).NotTo(Equal("1"))
			}
			Eventually(verifyObservedGeneration, 2*time.Minute).Should(Succeed())
		})
	})
})

// createSecret creates a Kubernetes secret with the given data.
func createSecret(namespace, name string, data map[string]string) {
	manifest := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
`, name, namespace)

	for key, value := range data {
		manifest += fmt.Sprintf("  %s: %q\n", key, value)
	}

	cmd := exec.Command("kubectl", "delete", "secret", name, "-n", namespace, "--ignore-not-found=true")
	_, _ = utils.Run(cmd)

	applyManifest(manifest)
}

// createSupabaseProject creates a SupabaseProject custom resource.
func createSupabaseProject(namespace, name, dbSecretName, storageSecretName string) {
	manifest := fmt.Sprintf(`
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: %s
  namespace: %s
spec:
  projectId: test-project
  database:
    secretRef:
      name: %s
  storage:
    secretRef:
      name: %s
`, name, namespace, dbSecretName, storageSecretName)

	applyManifest(manifest)
}

// applyManifest applies a Kubernetes manifest from a string.
func applyManifest(manifest string) {
	tmpFile, err := os.CreateTemp("", "manifest-*.yaml")
	Expect(err).NotTo(HaveOccurred())
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(manifest)
	Expect(err).NotTo(HaveOccurred())
	tmpFile.Close()

	cmd := exec.Command("kubectl", "apply", "-f", tmpFile.Name())
	_, err = utils.Run(cmd)
	Expect(err).NotTo(HaveOccurred())
}
