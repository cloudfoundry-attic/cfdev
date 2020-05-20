package daemon_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/daemon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var (
	winsw  *daemon.WinSW
	label  string
	tmpDir string
)

var _ = Describe("Winsw", func() {
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "testasset")
		Expect(err).To(BeNil())

		binaryDir := filepath.Join(tmpDir, "bin")

		err = os.MkdirAll(binaryDir, 0666)
		Expect(err).To(BeNil())

		assetPath := filepath.Join(binaryDir, "winsw.exe")

		err = downloadTestAsset(assetPath, "https://github.com/winsw/winsw/releases/download/v2.9.0/WinSW.NETCore31.x64.exe")
		Expect(err).To(BeNil())
		Expect(assetPath).To(BeAnExistingFile())
		winsw = daemon.NewWinSW(tmpDir)
		label = "some-daemon"
	})

	AfterEach(func() {
		winsw.Stop(label)
		winsw.RemoveDaemon(label)

		err := os.RemoveAll(tmpDir)
		Expect(err).To(BeNil())
	})

	Describe("AddDaemon", func() {
		It("should load the daemon", func() {
			spec := daemon.DaemonSpec{
				Label:            label,
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
			spec := daemon.DaemonSpec{
				Label:   label,
				Program: "powershell.exe",
			}

			Expect(winsw.AddDaemon(spec)).To(Succeed())
			output := getPowerShellOutput("get-service")
			Expect(output).To(ContainSubstring(label))

			Expect(winsw.RemoveDaemon(label)).To(Succeed())
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
			spec := daemon.DaemonSpec{
				Label:   label,
				Program: "powershell.exe",
				ProgramArguments: []string{
					fmt.Sprintf("'some-content' >> %s;", testFilePath),
					"Start-Sleep -Seconds 20",
				},
			}
			Expect(winsw.AddDaemon(spec)).To(Succeed())

			By("starting the service")
			Expect(winsw.Start(label)).To(Succeed())
			Eventually(func() bool {
				isRunning, _ := winsw.IsRunning(label)
				return isRunning
			}, 20, 1).Should(BeTrue())

			Eventually(testFilePath, 10*time.Second).Should(BeAnExistingFile())

			By("stopping the service")
			Expect(winsw.Stop(label)).To(Succeed())
			Eventually(func() bool {
				isRunning, _ := winsw.IsRunning(label)
				return isRunning
			}, 20, 1).Should(BeFalse())
		})
	})
})

func downloadTestAsset(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := netClient.Get(url)
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
