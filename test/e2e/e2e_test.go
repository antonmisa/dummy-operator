package e2e

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	//nolint:golint
	//nolint:revive
	. "github.com/onsi/ginkgo/v2"

	//nolint:golint
	//nolint:revive
	. "github.com/onsi/gomega"

	"github.com/antonmisa/dummy-operator/test/utils"
)

// constant parts of the file
const namespace = "dummy-operator-system"

var _ = Describe("dummy", Ordered, func() {
	BeforeAll(func() {
		// The namespace can be created when we run make install
		// However, in this test we want ensure that the solution
		// can run in a ns labeled as restricted. Therefore, we are
		// creating the namespace an lebeling it.
		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	Context("Dummy Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error
			projectDir, _ := utils.GetProjectDir()

			// operatorImage stores the name of the image used in the example
			var operatorImage = "interview.com/dummy-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", operatorImage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(operatorImage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", operatorImage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name
				cmd = exec.Command("kubectl", "get",
					"pods",
					"-o", "go-template='{{ range .items }}{{ if not .metadata.deletionTimestamp }}{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}'",
					"-n", namespace,
				)
				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) < 1 {
					return fmt.Errorf("expect controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

			By("creating an instance of the Dummy Operand(CR)")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "apply", "-f", filepath.Join(projectDir,
					"config/samples/_v1alpha1_dummy2.yaml"), "-n", namespace)
				_, err = utils.Run(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("validating that pod(s) status.phase=Running")
			getDummyPodStatus := func() error {
				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "app=dummy",
					"-o", "jsonpath={.items[*].status}", "-n", namespace,
				)
				status, err := utils.Run(cmd)
				fmt.Println(string(status))
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if !strings.Contains(string(status), "\"phase\":\"Running\"") {
					return fmt.Errorf("Dummy pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, getDummyPodStatus, time.Minute, time.Second).Should(Succeed())

			By("validating that the status of the custom resource created is updated or not")
			getStatus := func() error {
				// Validate specEcho
				cmd = exec.Command("kubectl", "get", "dummy",
					"dummy-sample", "-o", "jsonpath={.status.specEcho}",
					"-n", namespace,
				)
				echo, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(echo) != "I'm just a dummy-sample" {
					return fmt.Errorf("specEcho condition is not the same as message")
				}

				// Validate podStatus
				cmd = exec.Command("kubectl", "get", "dummy",
					"dummy-sample", "-o", "jsonpath={.status.podStatus}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if !strings.Contains(string(status), "Running") && !strings.Contains(string(status), "Pending") {
					return fmt.Errorf("podStatus condition with type Running/Pending should be set")
				}
				return nil
			}
			Eventually(getStatus, time.Minute, time.Second).Should(Succeed())
		})
	})
})
