package launchd_test

import (
	"code.cloudfoundry.org/cfdev/launchd"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	winsw    *launchd.WinSW
	label    string
	tmpDir   string
	assetDir string
)

var _ = Describe("launchd windows", func() {
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "testasset")
		Expect(err).To(BeNil())

		assetDir := filepath.Join(tmpDir, "cache")

		err = os.MkdirAll(filepath.Join(tmpDir, "cache"), 0666)
		Expect(err).To(BeNil())

		assetPath := filepath.Join(assetDir, "winsw.exe")

		err = downloadTestAsset(assetPath, "https://github.com/kohsuke/winsw/releases/download/winsw-v2.1.2/WinSW.NET4.exe")
		Expect(err).To(BeNil())
		Expect(assetPath).To(BeAnExistingFile())
		winsw = launchd.NewWinSW(tmpDir)
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpDir)
		Expect(err).To(BeNil())
	})

	Describe("launchd windows", func() {
		BeforeEach(func() {
			label = "some-daemon"
		})

		AfterEach(func() {
			spec := launchd.DaemonSpec{
				Label:     label,
				CfDevHome: tmpDir,
			}
			winsw.Stop(spec)
			winsw.RemoveDaemon(spec)
		})

		Describe("AddDaemon", func() {
			It("should load the daemon", func() {
				spec := launchd.DaemonSpec{
					Label:            label,
					CfDevHome:        tmpDir,
					Program:          "powershell.exe",
					ProgramArguments: []string{"echo 'hello'"},
				}

				Expect(winsw.AddDaemon(spec)).To(Succeed())
				output := getPowerShellOutput(fmt.Sprintf(`Get-Service | Where-Object { $_.Name -eq "%s" }`, label))

				Expect(output).NotTo(BeEmpty())
			})
		})

		Describe("RemoveDaemon", func() {
			It("should remove the daemon", func() {
				spec := launchd.DaemonSpec{
					Label:     label,
					CfDevHome: tmpDir,
					Program:   "powershell.exe",
				}

				Expect(winsw.AddDaemon(spec)).To(Succeed())
				output := getPowerShellOutput("get-service")
				Expect(output).To(ContainSubstring(label))

				Expect(winsw.RemoveDaemon(spec)).To(Succeed())
				output = getPowerShellOutput(fmt.Sprintf(`Get-Service | Where-Object { $_.Name -eq "%s" }`, label))
				Expect(output).To(BeEmpty())
			})
		})

		Describe("Lifecycle", func() {
			var testFilePath string

			BeforeEach(func() {
				testFilePath = filepath.Join(tmpDir, "test-file.txt")
			})

			AfterEach(func() {
				os.RemoveAll(testFilePath)
			})

			It("should start, stop the daemon", func() {
				By("adding the service")
				spec := launchd.DaemonSpec{
					Label:     label,
					CfDevHome: tmpDir,
					Program:   "powershell.exe",
					ProgramArguments: []string{
						fmt.Sprintf("'some-content' >> %s;", testFilePath),
						"Start-Sleep -Seconds 20",
					},
				}
				Expect(winsw.AddDaemon(spec)).To(Succeed())

				By("starting the service")
				Expect(winsw.Start(spec)).To(Succeed())
				Eventually(func() bool {
					isRunning, _ := winsw.IsRunning(spec)
					return isRunning
				}, 20, 1).Should(BeTrue())

				Eventually(testFilePath).Should(BeAnExistingFile())

				By("stopping the service")
				Expect(winsw.Stop(spec)).To(Succeed())
				Eventually(func() bool {
					isRunning, _ := winsw.IsRunning(spec)
					return isRunning
				}, 20, 1).Should(BeFalse())
			})
		})
	})
})

func downloadTestAsset(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getPowerShellOutput(command string) string {
	cmd := exec.Command("powershell.exe", "-Command", command)
	output, err := cmd.Output()
	Expect(err).NotTo(HaveOccurred())
	return strings.TrimSpace(string(output))
}
