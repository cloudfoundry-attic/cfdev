package launchd_test

import (
	"code.cloudfoundry.org/cfdevd/launchd"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
	"os/exec"
	"strings"
	"path/filepath"
	"os"
)

var _ = Describe("launchd windows", func() {

	var (
		lnchd launchd.Launchd
		label  string
	)

	BeforeEach(func() {
		lnchd = launchd.Launchd{}
		label = randomDaemonName()
	})

	AfterEach(func() {
		lnchd.Stop(label)
		lnchd.RemoveDaemon(label)
	})

	Describe("AddDaemon", func() {
		It("should load the daemon", func() {
			spec := launchd.DaemonSpec{
				Label: label,
			}

			Expect(lnchd.AddDaemon(spec)).To(Succeed())

			output := getPowerShellOutput(fmt.Sprintf(`Get-Service | Where-Object { $_.Name -eq "%s" }`, label))
			Expect(output).NotTo(BeEmpty())
		})
	})

	Describe("RemoveDaemon", func() {
		It("should remove the daemon", func() {
			spec := launchd.DaemonSpec{
				Label: label,
			}

			Expect(lnchd.AddDaemon(spec)).To(Succeed())
			output := getPowerShellOutput("get-service")
			Expect(output).To(ContainSubstring(label))

			Expect(lnchd.RemoveDaemon(label)).To(Succeed())
			output = getPowerShellOutput(fmt.Sprintf(`Get-Service | Where-Object { $_.Name -eq "%s" }`, label))
			Expect(output).To(BeEmpty())
		})
	})

	Describe("Lifecycle", func() {

		var tempDir string

		BeforeEach(func() {
			tempDir = os.Getenv("TEMP")
		})

		AfterEach(func() {
			os.RemoveAll(filepath.Join(tempDir, "simple-file.txt"))
		})

		FIt("should start, stop the daemon", func() {
			By("adding the service")
			spec := launchd.DaemonSpec{
				Label: label,
				Program: "powershell.exe",
				ProgramArguments: []string{
					"-Command",
					`"some-content" >> $env:TEMP\simple-file.txt`,
				},
			}
			Expect(lnchd.AddDaemon(spec)).To(Succeed())

			By("starting the service")
			Expect(lnchd.Start(spec)).To(Succeed())
			//Expect(filepath.Join(tempDir, "simple-file.txt")).To(BeAnExistingFile())
			//Eventually(func() string {
			//	return filepath.Join(tempDir, "simple-file.txt")
			//}, 60, 1).Should(BeAnExistingFile())
			Eventually(filepath.Join(tempDir, "simple-file.txt")).Should(BeAnExistingFile())
			Expect(lnchd.IsRunning(label)).To(BeTrue())

			By("stopping the service")
			Expect(lnchd.Stop(label)).To(Succeed())
			Expect(lnchd.IsRunning(label)).To(BeFalse())
		})
	})
})

func getPowerShellOutput(command string) string {
	cmd := exec.Command("powershell.exe", "-Command", command)
	output, err := cmd.Output()
	Expect(err).NotTo(HaveOccurred())
	return strings.TrimSpace(string(output))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomDaemonName() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return "some-daemon" + string(b)
}
