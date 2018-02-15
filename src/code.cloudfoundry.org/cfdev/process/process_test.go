package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/process"
	"os/exec"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"path/filepath"
	"os"
	"strconv"
)

var _ = Describe("Process Test", func() {
	Describe("Terminate", func() {
		It("", func() {
			Expect(true).To(BeTrue())
		})

		Context("when the pidfile contains a valid pid", func() {
			It("sends a SIGTERM to the process with the given pid from the given pidfile", func() {
				cmd := exec.Command("sleep", "99999")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				pidfile := filepath.Join(os.Getenv("TMPDIR"), "some-pidfile")
				ioutil.WriteFile(pidfile, []byte(strconv.Itoa(cmd.Process.Pid)), os.ModePerm)

				Expect(process.Terminate(pidfile)).To(Succeed())
				Eventually(session).Should(gexec.Exit())
				Expect(session.ExitCode()).NotTo(Equal(0))
			})
		})

		Context("when the pidfile cannot be read", func() {
			It("returns an error", func() {
				Expect(process.Terminate("some-bad-pidfile")).To(MatchError("failed to read pidfile some-bad-pidfile"))
			})
		})

		Context("when the pidfile does not contain a pid", func() {
			It("returns an error", func() {
				pidfile := filepath.Join(os.Getenv("TMPDIR"), "some-bad-pidfile")
				ioutil.WriteFile(pidfile, []byte("some-bad-pid"), os.ModePerm)
				Expect(process.Terminate(pidfile).Error()).To(ContainSubstring("some-bad-pidfile did not contain an integer"))
			})
		})
	})

	Describe("Kill", func() {
		It("", func() {
			Expect(true).To(BeTrue())
		})

		Context("when the pidfile contains a valid pid", func() {
			It("sends a SIGTERM to the process with the given pid from the given pidfile", func() {
				cmd := exec.Command("sleep", "99999")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				pidfile := filepath.Join(os.Getenv("TMPDIR"), "some-pidfile")
				ioutil.WriteFile(pidfile, []byte(strconv.Itoa(cmd.Process.Pid)), os.ModePerm)

				Expect(process.Kill(pidfile)).To(Succeed())
				Eventually(session).Should(gexec.Exit())
				Expect(session.ExitCode()).NotTo(Equal(0))
			})
		})

		Context("when the pidfile cannot be read", func() {
			It("returns an error", func() {
				Expect(process.Kill("some-bad-pidfile")).To(MatchError("failed to read pidfile some-bad-pidfile"))
			})
		})

		Context("when the pidfile does not contain a pid", func() {
			It("returns an error", func() {
				pidfile := filepath.Join(os.Getenv("TMPDIR"), "some-bad-pidfile")
				ioutil.WriteFile(pidfile, []byte("some-bad-pid"), os.ModePerm)
				Expect(process.Kill(pidfile).Error()).To(ContainSubstring("some-bad-pidfile did not contain an integer"))
			})
		})
	})
})
