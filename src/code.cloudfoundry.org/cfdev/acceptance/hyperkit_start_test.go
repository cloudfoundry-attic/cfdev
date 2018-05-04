package acceptance

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cfdevd/launchd"
	"code.cloudfoundry.org/cfdevd/launchd/models"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit start", func() {
	var (
		stateDir string
		cacheDir string
	)

	BeforeEach(func() {
		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")

		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(cacheDir)
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}
	})

	AfterEach(func() {
		gexec.Kill()
		os.Unsetenv("CFDEV_SKIP_ASSET_CHECK")
	})

	Context("when lacking sudo privileges", func() {
		BeforeEach(func(done Done) {
			Expect(HasSudoPrivilege()).To(BeFalse())
			close(done)
		}, 10)

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

	Context("cfdev linuxkit is already running", func() {
		var tmpDir string
		var originalPid string
		BeforeEach(func() {
			tmpDir, _ = ioutil.TempDir("", "cfdev.test.running.")
			lctl := launchd.New(tmpDir)
			Expect(lctl.IsRunning("org.cloudfoundry.cfdev.linuxkit")).To(BeFalse())
			lctl.AddDaemon(models.DaemonSpec{
				Label:            "org.cloudfoundry.cfdev.linuxkit",
				Program:          "/bin/bash",
				ProgramArguments: []string{"/bin/bash", "-c", "sleep 300"},
				RunAtLoad:        true,
			})
			Eventually(func() (bool, error) { return lctl.IsRunning("org.cloudfoundry.cfdev.linuxkit") }).Should(BeTrue())
			originalPid, _ = LaunchdPid("org.cloudfoundry.cfdev.linuxkit")
			Expect(originalPid).NotTo(BeEmpty())
		})

		AfterEach(func() {
			exec.Command("launchctl", "unload", "-w", filepath.Join(tmpDir, "org.cloudfoundry.cfdev.linuxkit.plist")).Run()
			os.RemoveAll(tmpDir)
		})

		It("doesn't restart the linuxkit process", func() {
			session := cf.Cf("dev", "start")
			Eventually(session).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say("CF Dev is already running..."))
			Expect(LaunchdPid("org.cloudfoundry.cfdev.linuxkit")).To(Equal(originalPid))
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

func LaunchdPid(label string) (string, error) {
	out, err := exec.Command("launchctl", "list", label).Output()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`"PID" = (\d+);`)
	for _, line := range strings.Split(string(out), "\n") {
		results := re.FindStringSubmatch(line)
		if len(results) > 0 {
			return results[1], nil
		}
	}
	return "", fmt.Errorf("PID not found")
}
