package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var _ = Describe("ServiceWrapper Lifecycle", func() {

	var (
		tempDir string
		servicewPath string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cfdev-service-wrapper-")
		Expect(err).NotTo(HaveOccurred())

		servicewPath = filepath.Join(tempDir, "test")
		servicewConfigPath := filepath.Join(tempDir, "test.yml")

		copy(binaryPath, servicewPath, true)
		copy(fixturePath("simple.yml"), servicewConfigPath)
	})

	AfterEach(func() {
		exec.Command(servicewPath, "stop").Run()
		exec.Command(servicewPath, "uninstall").Run()
		os.RemoveAll(tempDir)
	})

	It("installs, runs, and remove services", func() {
		Expect(status(servicewPath)).To(ContainSubstring("Error"))

		run(servicewPath, "install")
		Expect(status(servicewPath)).To(Equal("Stopped"))

		run(servicewPath, "start")
		time.Sleep(3*time.Second)
		Expect(status(servicewPath)).To(Equal("Running"))

		if runtime.GOOS != "windows" {
			output := run("bash", "-c", "ps aux | grep 'sleep 12345'")
			Expect(strings.TrimSpace(output)).NotTo(BeEmpty())
		}

		run(servicewPath, "stop")
		Expect(status(servicewPath)).To(Equal("Stopped"))

		run(servicewPath, "uninstall")
		Expect(status(servicewPath)).To(ContainSubstring("Error"))
	})
})

func status(servicewPath string) string {
	return strings.TrimSpace(run(servicewPath, "status"))
}