//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/strrl/supabase-operator/test/utils"
)

// namespace where the project is deployed in
const namespace = "supabase-operator-system"
const helmReleaseName = "supabase-operator"

var _ = Describe("Manager", Ordered, func() {
	var controllerPodName string

	// Before running the tests, set up the environment by creating the namespace,
	// enforce the restricted security policy to the namespace, and installing the controller via Helm.
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

		By("installing the operator via Helm chart")
		cmd = exec.Command("helm", "upgrade", "--install", helmReleaseName, "helm/supabase-operator",
			"--namespace", namespace,
			"--create-namespace",
			"--wait",
			"--timeout", "5m",
			"--set", fmt.Sprintf("image.repository=%s", projectImageRepository),
			"--set", fmt.Sprintf("image.tag=%s", projectImageTag),
			"--set", "image.pullPolicy=IfNotPresent",
		)
		_, err = utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred(), "Failed to install the operator Helm release")
	})

	// After all tests have been executed, clean up by uninstalling the Helm release, removing CRDs,
	// and deleting the namespace.
	AfterAll(func() {
		By("uninstalling the operator Helm release")
		cmd := exec.Command("helm", "uninstall", helmReleaseName, "-n", namespace, "--wait")
		_, _ = utils.Run(cmd)

		By("cleaning up Supabase CRDs")
		cmd = exec.Command("kubectl", "delete", "crd", "supabaseprojects.supabase.strrl.dev", "--ignore-not-found=true")
		_, _ = utils.Run(cmd)

		By("removing manager namespace")
		cmd = exec.Command("kubectl", "delete", "ns", namespace, "--ignore-not-found=true")
		_, _ = utils.Run(cmd)
	})

	// After each test, check for failures and collect logs, events,
	// and pod descriptions for debugging.
	AfterEach(func() {
		specReport := CurrentSpecReport()
		if specReport.Failed() {
			if controllerPodName == "" {
				_, _ = fmt.Fprintf(GinkgoWriter, "controllerPodName not set; skipping pod log/describe collection\n")
				return
			}
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
					"pods",
					"-l", "app.kubernetes.io/name=supabase-operator",
					"-l", fmt.Sprintf("app.kubernetes.io/instance=%s", helmReleaseName),
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
				g.Expect(controllerPodName).To(ContainSubstring(helmReleaseName))

				// Validate the pod's status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"), "Incorrect controller-manager pod status")

				// Wait for pod to be ready (especially important for webhooks)
				cmd = exec.Command("kubectl", "wait",
					"pod", controllerPodName,
					"--for=condition=Ready",
					"--timeout=60s",
					"-n", namespace,
				)
				_, err = utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred(), "Pod did not become ready")
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
		const postgresName = "test-postgres"

		BeforeAll(func() {
			By("deploying PostgreSQL for testing")
			deployPostgres(testNamespace, postgresName)

			By("waiting for PostgreSQL to be ready")
			waitForPostgres(testNamespace, postgresName)
		})

		BeforeEach(func() {
			By("creating database secret")
			createSecret(testNamespace, dbSecretName, map[string]string{
				"host":     fmt.Sprintf("%s.%s.svc.cluster.local", postgresName, testNamespace),
				"port":     "5432",
				"database": "supabase",
				"username": "postgres",
				"password": "testpassword",
			})

			By("creating storage secret")
			createSecret(testNamespace, storageSecretName, map[string]string{
				"endpoint":        "s3.amazonaws.com",
				"region":          "us-east-1",
				"bucket":          "test-bucket",
				"accessKeyId":     "test-access-key",
				"secretAccessKey": "test-secret-key",
			})
		})

		AfterAll(func() {
			By("cleaning up SupabaseProject")
			cmd := exec.Command("kubectl", "delete", "supabaseproject", projectName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)

			By("cleaning up secrets")
			cmd = exec.Command("kubectl", "delete", "secret", dbSecretName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)
			cmd = exec.Command("kubectl", "delete", "secret", storageSecretName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)

			By("cleaning up PostgreSQL")
			cmd = exec.Command("kubectl", "delete", "deployment", postgresName, "-n", testNamespace, "--ignore-not-found=true")
			_, _ = utils.Run(cmd)
			cmd = exec.Command("kubectl", "delete", "service", postgresName, "-n", testNamespace, "--ignore-not-found=true")
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

		It("should validate secret references via webhook (T105a)", func() {
			By("attempting to create SupabaseProject with empty database secret reference")
			manifest := fmt.Sprintf(`
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: %s-empty-ref
  namespace: %s
spec:
  projectId: test-project-empty-ref
  database:
    secretRef:
      name: ""
  storage:
    secretRef:
      name: %s
`, projectName, testNamespace, storageSecretName)

			tmpFile, err := os.CreateTemp("", "manifest-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(manifest)
			Expect(err).NotTo(HaveOccurred())
			tmpFile.Close()

			By("expecting webhook to reject the request")
			cmd := exec.Command("kubectl", "apply", "-f", tmpFile.Name())
			output, err := utils.Run(cmd)
			Expect(err).To(HaveOccurred(), "Expected webhook to reject empty secret reference")
			Expect(output).To(ContainSubstring("admission webhook"))
			Expect(output).To(Or(
				ContainSubstring("cannot be empty"),
				ContainSubstring("secretRef.name"),
			))
		})

		It("should reject database secrets missing required keys (T105b)", func() {
			By("creating invalid database secret (missing required keys)")
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

			tmpFile, err := os.CreateTemp("", "manifest-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(manifest)
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile.Close()).To(Succeed())

			By("expecting admission webhook to reject the SupabaseProject")
			cmd := exec.Command("kubectl", "apply", "-f", tmpFile.Name())
			output, err := utils.Run(cmd)
			Expect(err).To(HaveOccurred(), "Expected webhook to reject SupabaseProject with invalid database secret")
			Expect(output).To(ContainSubstring("admission webhook"))
			Expect(output).To(ContainSubstring("database secret missing required key"))

			By("cleaning up invalid resources")
			cmd = exec.Command("kubectl", "delete", "secret", "invalid-db-secret", "-n", testNamespace, "--ignore-not-found=true")
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

		It("should capture Kong Studio screenshot via Basic Auth (T107)", func() {
			const (
				localStudioPort = 18080
				studioUser      = "supabase"
				studioPassword  = "this_password_is_insecure_and_should_be_updated"
			)

			By("creating a SupabaseProject CR")
			createSupabaseProject(testNamespace, projectName, dbSecretName, storageSecretName)

			By("waiting for SupabaseProject to reach Running phase")
			verifyRunning := func(g Gomega) {
				cmd := exec.Command("kubectl", "get", "supabaseproject", projectName, "-n", testNamespace,
					"-o", "jsonpath={.status.phase}")
				output, err := utils.Run(cmd)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(output).To(Equal("Running"))
			}
			Eventually(verifyRunning, 3*time.Minute).Should(Succeed())

			By("waiting for Kong pod to report Ready")
			var kongPodName string
			Eventually(func() error {
				cmd := exec.Command("kubectl", "get", "pods",
					"-l", fmt.Sprintf("app.kubernetes.io/instance=%s", projectName),
					"-l", "app.kubernetes.io/name=kong",
					"-n", testNamespace,
					"-o", "json")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				var podList struct {
					Items []struct {
						Metadata struct {
							Name string `json:"name"`
						} `json:"metadata"`
						Status struct {
							ContainerStatuses []struct {
								Ready bool `json:"ready"`
							} `json:"containerStatuses"`
						} `json:"status"`
					} `json:"items"`
				}
				if err := json.Unmarshal([]byte(output), &podList); err != nil {
					return err
				}
				for _, item := range podList.Items {
					ready := false
					for _, cs := range item.Status.ContainerStatuses {
						if cs.Ready {
							ready = true
							break
						}
					}
					if ready {
						kongPodName = item.Metadata.Name
						return nil
					}
				}
				return fmt.Errorf("kong pod not ready yet")
			}, 4*time.Minute, 2*time.Second).Should(Succeed())

			By("waiting for Kong service to have ready endpoints")
			Eventually(func() error {
				// Check if the Kong service has ready endpoints
				cmd := exec.Command("kubectl", "get", "endpoints",
					fmt.Sprintf("%s-kong", projectName),
					"-n", testNamespace,
					"-o", "jsonpath={.subsets[*].addresses[*].ip}")
				output, err := utils.Run(cmd)
				if err != nil {
					return err
				}
				if output == "" {
					return fmt.Errorf("kong service has no ready endpoints")
				}
				return nil
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("waiting additional time for Kong to fully initialize")
			time.Sleep(10 * time.Second)

			By("port-forwarding the Kong pod to localhost")
			portForwardCmd, err := startPortForward(testNamespace, fmt.Sprintf("pod/%s", kongPodName), localStudioPort, 8000)
			Expect(err).NotTo(HaveOccurred())
			defer stopPortForward(portForwardCmd)

			targetURL := fmt.Sprintf("http://127.0.0.1:%d/project/default", localStudioPort)
			client := &http.Client{Timeout: 5 * time.Second}
			Eventually(func() error {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
					targetURL, nil)
				if err != nil {
					return err
				}
				req.SetBasicAuth(studioUser, studioPassword)
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("unexpected status %d", resp.StatusCode)
				}
				return nil
			}, 2*time.Minute, 2*time.Second).Should(Succeed())

			chromePath, err := utils.FindChromeBinary()
			if err != nil {
				Skip(fmt.Sprintf("headless chrome not available: %v", err))
			}

			By("capturing the Kong Studio page screenshot")
			timestamp := time.Now().Format("20060102-150405")
			screenshotPath := filepath.Join(".artifacts", "screenshots", fmt.Sprintf("kong-studio-%s.png", timestamp))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			err = captureStudioScreenshot(ctx, chromePath,
				targetURL,
				studioUser, studioPassword, screenshotPath)
			Expect(err).NotTo(HaveOccurred())
			_, _ = fmt.Fprintf(GinkgoWriter, "Studio screenshot saved to %s\n", screenshotPath)
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

func startPortForward(namespace, resource string, localPort, remotePort int) (*exec.Cmd, error) {
	cmd := exec.Command("kubectl", "port-forward",
		resource,
		fmt.Sprintf("%d:%d", localPort, remotePort),
		"-n", namespace,
	)
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			stopPortForward(cmd)
			return nil, fmt.Errorf("port-forward did not become ready: %w", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	return cmd, nil
}

func stopPortForward(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_, _ = fmt.Fprintf(GinkgoWriter, "Waiting for port-forward to exit...\n")
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_, _ = fmt.Fprintf(GinkgoWriter, "Port-forward did not exit gracefully, killing...\n")
		_ = cmd.Process.Kill()
		<-done
	}
}

func captureStudioScreenshot(ctx context.Context, chromePath, url, username, password, outputPath string) error {
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("window-size", "1280,720"),
		chromedp.Flag("remote-allow-origins", "*"),
	)

	var (
		screenshot []byte
		lastErr    error
	)

	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptAllocatorCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
		attemptCtx, cancelTimeout := context.WithTimeout(attemptAllocatorCtx, 5*time.Minute)

		browserCtx, cancelCtx := chromedp.NewContext(
			attemptCtx,
			chromedp.WithLogf(func(format string, a ...any) {
				_, _ = fmt.Fprintf(GinkgoWriter, "[chromedp] "+format+"\n", a...)
			}),
		)

		var attemptScreenshot []byte
		tasks := chromedp.Tasks{
			network.Enable(),
			network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
				"Authorization": authHeader,
			})),
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(2 * time.Second),
			chromedp.CaptureScreenshot(&attemptScreenshot),
		}

		runErr := chromedp.Run(browserCtx, tasks)
		cancelCtx()
		cancelTimeout()
		cancelAlloc()

		if runErr == nil {
			screenshot = attemptScreenshot
			lastErr = nil
			break
		}

		lastErr = runErr
		_, _ = fmt.Fprintf(GinkgoWriter, "chromedp attempt %d failed: %v\n", attempt, runErr)

		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if lastErr != nil {
		return fmt.Errorf("capture studio screenshot failed after %d attempts: %w", maxAttempts, lastErr)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, screenshot, 0o644)
}

// deployPostgres deploys a PostgreSQL instance for testing.
func deployPostgres(namespace, name string) {
	// Create Deployment
	deploymentManifest := fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_PASSWORD
          value: testpassword
        - name: POSTGRES_DB
          value: supabase
        ports:
        - containerPort: 5432
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - postgres
          initialDelaySeconds: 5
          periodSeconds: 5
`, name, namespace, name, name)

	applyManifest(deploymentManifest)

	// Create Service
	serviceManifest := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  selector:
    app: %s
  ports:
  - port: 5432
    targetPort: 5432
`, name, namespace, name)

	applyManifest(serviceManifest)
}

// waitForPostgres waits for PostgreSQL to be ready.
func waitForPostgres(namespace, name string) {
	Eventually(func(g Gomega) {
		cmd := exec.Command("kubectl", "get", "deployment", name, "-n", namespace,
			"-o", "jsonpath={.status.readyReplicas}")
		output, err := utils.Run(cmd)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(output).To(Equal("1"))
	}, 3*time.Minute, 5*time.Second).Should(Succeed())
}
