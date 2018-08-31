// +build darwin

package hypervisor_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cfdev/hypervisor"
)

var _ = Describe("safe kill test", func() {
	var (
		err           error
		processToKill *gexec.Session
		tmpDir        string
		pidFile       string
	)

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir(os.Getenv("TMPDIR"), "safe-kill-test")
		Expect(err).NotTo(HaveOccurred())
		processToKill, err = gexec.Start(exec.Command("sleep", "36000"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		pidFile = filepath.Join(tmpDir, "processToKill.pid")
		ioutil.WriteFile(
			pidFile,
			[]byte(strconv.Itoa(processToKill.Command.Process.Pid)),
			0644,
		)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
		gexec.KillAndWait()
	})

	Context("process is still running", func() {
		It("kills the process and clean up the pidfile", func() {
			Expect(hypervisor.SafeKill(pidFile, "sleep")).To(Succeed())
			Eventually(processToKill).Should(gexec.Exit())
			Expect(pidFile).NotTo(BeAnExistingFile())
		})
	})

	Context("process is still running with different filename", func() {
		It("leaves process running and cleans up the pidfile", func() {
			Expect(hypervisor.SafeKill(pidFile, "other")).To(Succeed())
			Expect(pidFile).NotTo(BeAnExistingFile())
			Expect(processToKill).ShouldNot(gexec.Exit())
		})
	})

	Context("process is no longer running", func() {
		BeforeEach(func() {
			gexec.KillAndWait()
			Expect(processToKill).To(gexec.Exit())
		})

		It("cleans up the pidfile", func() {
			Expect(hypervisor.SafeKill(pidFile, "sleep")).To(Succeed())
			Eventually(processToKill).Should(gexec.Exit())
			Expect(pidFile).NotTo(BeAnExistingFile())
		})
	})
})
