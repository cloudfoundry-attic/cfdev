package acceptance

import (
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
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
		cfHome, err := ioutil.TempDir("", "cf-home")
		Expect(err).ToNot(HaveOccurred())
		cfdevHome = CreateTempCFDevHomeDir()
		os.Setenv("CF_HOME", cfHome)
		os.Setenv("CFDEV_HOME", cfdevHome)
		session := cf.Cf("install-plugin", pluginPath, "-f")
		Eventually(session).Should(gexec.Exit(0))
		session = cf.Cf("plugins")
		Eventually(session).Should(gbytes.Say("cfdev"))
		Eventually(session).Should(gexec.Exit(0))

		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")
		linuxkitPidPath = filepath.Join(stateDir, "linuxkit.pid")

		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(cacheDir)
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}
	})

	AfterEach(func() {
		gexec.Kill()
		pid := PidFromFile(linuxkitPidPath)

		if pid != 0 {
			syscall.Kill(int(-pid), syscall.SIGKILL)
		}

		os.RemoveAll(cfdevHome)
		session := cf.Cf("uninstall-plugin", "cfdev")
		Eventually(session).Should(gexec.Exit(0))

		os.Unsetenv("CF_HOME")
		os.Unsetenv("CFDEV_HOME")
		os.Unsetenv("CFDEV_SKIP_ASSET_CHECK")
	})

	Context("when lacking sudo privileges", func() {
		BeforeEach(func() {
			Expect(HasSudoPrivilege()).To(BeFalse())
		})

		Context("BOSH & CF Router IP addresses are not aliased", func() {
			BeforeEach(func() {
				ExpectIPAddressedToNotBeAliased(BoshDirectorIP, CFRouterIP)
			})

			It("notifies and prompts for a password in order to sudo", func() {
				session := cf.Cf("dev", "start")
				Eventually(session.Out).Should(gbytes.Say("Setting up IP aliases"))
				Eventually(session.Err).Should(gbytes.Say("Password:"))
			})
		})
	})

	Context("with an unsupported distribution", func() {
		It("exits with code 1", func() {
			session := cf.Cf("dev", "start", "-f", "UNSUPPORTTED")
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with an unsupported version", func() {
		It("exits with code 1", func() {
			session := cf.Cf("dev", "start", "-n", "9.9.9.9.9")
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

		It("exits with code 1", func() {
			session := cf.Cf("dev", "start")
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
			cf.Cf("dev", "start")
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
			session := cf.Cf("dev", "start")
			Eventually(session).Should(gexec.Exit(0))

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
