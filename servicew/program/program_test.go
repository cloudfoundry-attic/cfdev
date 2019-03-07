package program_test

import (
	"code.cloudfoundry.org/cfdev/servicew/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "code.cloudfoundry.org/cfdev/servicew/program"
)

var _ = Describe("Program", func() {
	var (
		p       *Program
		tempDir string
	)

	BeforeEach(func() {
		tempDir = ""
	})

	AfterEach(func() {
		// sleep just in case
		// to wait for process
		// to actually start
		time.Sleep(time.Second)
		p.Stop(nil)

		os.RemoveAll(tempDir)
	})

	It("starts a process with a standard config", func() {
		var err error
		p, err = New(config.Config{
			Label:      "cfdev-program-test",
			Executable: "sleep",
			Args:       []string{"1234"},
		}, ioutil.Discard)
		Expect(err).NotTo(HaveOccurred())

		err = p.Start(nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() string {
			output, _ := exec.Command("ps", "aux").Output()
			return string(output)
		}, 10*time.Second).Should(ContainSubstring("sleep 1234"))
	})

	It("starts a process with environment variables", func() {
		var err error
		p, err = New(config.Config{
			Label:      "cfdev-program-test",
			Executable: "sh",
			Args:       []string{"-c", "sleep $SLEEP_COUNT"},
			Env: map[string]string{
				"SLEEP_COUNT": "1235",
			},
		}, ioutil.Discard)
		Expect(err).NotTo(HaveOccurred())

		err = p.Start(nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() string {
			output, _ := exec.Command("ps", "aux").Output()
			return string(output)
		}, 10*time.Second).Should(ContainSubstring("sleep 1235"))
	})

	It("starts a process with a log file specified", func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cfdev-program-test-")
		Expect(err).NotTo(HaveOccurred())

		logPath := filepath.Join(tempDir, "program-test.log")

		p, err = New(config.Config{
			Label:      "cfdev-program-test",
			Executable: "sh",
			Args:       []string{"-c", "while true; do echo hello; sleep 1; done"},
			Log:        logPath,
		}, ioutil.Discard)
		Expect(err).NotTo(HaveOccurred())

		err = p.Start(nil)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() string {
			output, _ := ioutil.ReadFile(logPath)
			return string(output)
		}, 10*time.Second).Should(ContainSubstring("hello"))
	})
})
