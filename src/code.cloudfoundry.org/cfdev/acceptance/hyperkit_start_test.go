package acceptance

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit start", func() {

	var (
		cfdevHome       string
		linuxkitPidPath string
		stateDir        string
		cacheDir        string
	)

	BeforeEach(func() {
		cfdevHome = CreateTempCFDevHomeDir()
		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")
		linuxkitPidPath = filepath.Join(stateDir, "linuxkit.pid")

		SetupDependencies(cacheDir)
	})

	AfterEach(func() {
		gexec.KillAndWait()
		pid := PidFromFile("linuxkit.pid")

		if pid != 0 {
			syscall.Kill(int(-pid), syscall.SIGKILL)
		}

		os.RemoveAll(cfdevHome)
	})

	Context("when not running as root", func() {
		BeforeEach(func() {
			me, err := user.Current()
			Expect(err).ToNot(HaveOccurred())
			Expect(me.Uid).ToNot(Equal("0"), "test should not run as root")
		})

		Context("BOSH & CF Router IP addresses are not aliased", func() {
			BeforeEach(func() {
				ExpectIPAddressedToNotBeAliased(BoshDirectorIP, CFRouterIP)
			})

			It("exits with a code 1", func() {
				command := exec.Command(cliPath, "start")
				command.Env = append(os.Environ(),
					fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))
				session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)

				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})

	Context("with an unsupported distribution", func() {
		It("exits with code 1", func() {
			command := exec.Command(cliPath, "start", "-f", "UNSUPPORTTED")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))
			session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with an unsupported version", func() {
		It("exits with code 1", func() {
			command := exec.Command(cliPath, "start", "-n", "9.9.9.9.9")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))
			session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when CFDEV_HOME is not writable", func() {
		BeforeEach(func() {
			os.Chmod(cfdevHome, 0555)
		})

		AfterEach(func() {
			os.Chmod(cfdevHome, 0777)
		})

		It("fails to start linuxkit", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(linuxkitPidPath).ShouldNot(BeAnExistingFile())
		})
	})

	Context("when the state directory has existing files", func() {
		var dirtyFile string

		BeforeEach(func() {
			err := os.MkdirAll(stateDir, 0777)
			Expect(err).ToNot(HaveOccurred())

			dirtyFile = filepath.Join(stateDir, "dirty")
			err = ioutil.WriteFile(dirtyFile, []byte{}, 0777)
			Expect(err).ToNot(HaveOccurred())
		})

		It("recreates a clean state directory", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(dirtyFile, 10, 1).ShouldNot(BeAnExistingFile())
		})
	})

	Context("the linuxkit pid file references an existing process", func() {
		var (
			existingCmd *exec.Cmd
			existingPid int
			exited      int32
		)

		BeforeEach(func() {
			err := os.MkdirAll(stateDir, 0777)
			Expect(err).ToNot(HaveOccurred())

			existingCmd = exec.Command("sleep", "300")
			err = existingCmd.Start()
			Expect(err).ToNot(HaveOccurred())

			existingPid = existingCmd.Process.Pid
			err = ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(existingPid)), 0777)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				existingCmd.Wait()
				atomic.StoreInt32(&exited, 1)
			}()
		})

		AfterEach(func() {
			existingCmd.Process.Kill()
		})

		It("doesn't restart the linuxkit process", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, 10, 1).Should(gexec.Exit(0))

			Expect(PidFromFile(linuxkitPidPath)).To(Equal(existingPid))
			Expect(atomic.LoadInt32(&exited)).To(BeEquivalentTo(0))
		})
	})
})

func ExpectIPAddressedToNotBeAliased(aliases ...string) {
	addrs, err := net.InterfaceAddrs()
	Expect(err).ToNot(HaveOccurred())

	for _, addr := range addrs {
		for _, alias := range aliases {
			Expect(addr.String()).ToNot(Equal(alias + "/32"))
		}
	}
}
