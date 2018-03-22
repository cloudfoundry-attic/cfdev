package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"syscall"

	"code.cloudfoundry.org/cfdev/process"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Process Test", func() {
	Describe("SignalAndCleanup", func() {
		var pidfile string
		BeforeEach(func() {
			f, _ := ioutil.TempFile("", "pidfile.")
			pidfile = f.Name()
			f.Close()
		})
		AfterEach(func() {
			os.Remove(pidfile)
		})

		It("", func() {
			Expect(true).To(BeTrue())
		})

		Context("when the pidfile contains a valid pid", func() {
			It("sends a SIGTERM to the process with the given pid from the given pidfile", func() {
				cmd := exec.Command("sleep", "99999")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				ioutil.WriteFile(pidfile, []byte(strconv.Itoa(cmd.Process.Pid)), os.ModePerm)

				Expect(process.SignalAndCleanup(pidfile, "sleep", syscall.SIGTERM)).To(Succeed())
				Eventually(session).Should(gexec.Exit())
				Expect(session.ExitCode()).NotTo(Equal(0))
			})
		})

		Context("when the pidfile cannot be read", func() {
			BeforeEach(func() {
				os.Chmod(pidfile, 000)
			})
			It("returns an error", func() {
				Expect(process.SignalAndCleanup(pidfile, "sleep", syscall.SIGTERM)).To(MatchError("failed to read pidfile " + pidfile))
			})
		})

		Context("when the pidfile does not contain a pid", func() {
			It("returns an error", func() {
				ioutil.WriteFile(pidfile, []byte("some-bad-pid"), os.ModePerm)
				Expect(process.SignalAndCleanup(pidfile, "sleep", syscall.SIGTERM).Error()).To(ContainSubstring(pidfile + " did not contain an integer"))
			})
		})

		Context("process description does not contain matcher", func() {
			var session *gexec.Session
			BeforeEach(func() {
				var err error
				cmd := exec.Command("sleep", "99999")
				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				ioutil.WriteFile(pidfile, []byte(strconv.Itoa(cmd.Process.Pid)), os.ModePerm)
			})

			AfterEach(func() {
				session.Kill()
			})

			It("leaves process running and deletes pid file", func() {
				Expect(process.SignalAndCleanup(pidfile, "NOT_A_MATCH", syscall.SIGTERM)).To(Succeed())
				Expect(pidfile).ToNot(BeAnExistingFile())
				Expect(session.ExitCode()).To(Equal(-1))
			})
		})
	})
})
