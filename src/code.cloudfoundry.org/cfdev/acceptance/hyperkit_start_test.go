package acceptance

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cfdevd/launchd"
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
		stateDir = filepath.Join(cfdevHome, "state", "linuxkit")

		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(cacheDir)
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}
	})

	AfterEach(func() {
		gexec.Kill()
		os.Unsetenv("CFDEV_SKIP_ASSET_CHECK")
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
		var launchdTmpDir string
		var originalPid string
		BeforeEach(func() {
			launchdTmpDir, _ = ioutil.TempDir("", "cfdev.test.running.")
			lctl := launchd.New(launchdTmpDir)
			Expect(lctl.IsRunning("org.cloudfoundry.cfdev.linuxkit")).To(BeFalse())
			lctl.AddDaemon(launchd.DaemonSpec{
				Label:            "org.cloudfoundry.cfdev.linuxkit",
				Program:          "/bin/bash",
				SessionType:      "Background",
				ProgramArguments: []string{"/bin/bash", "-c", "sleep 300"},
				RunAtLoad:        true,
			})
			Eventually(func() (bool, error) { return lctl.IsRunning("org.cloudfoundry.cfdev.linuxkit") }).Should(BeTrue())
			originalPid, _ = LaunchdPid("org.cloudfoundry.cfdev.linuxkit")
			Expect(originalPid).NotTo(BeEmpty())
		})

		AfterEach(func() {
			exec.Command("launchctl", "unload", "-w", filepath.Join(launchdTmpDir, "org.cloudfoundry.cfdev.linuxkit.plist")).Run()
			os.RemoveAll(launchdTmpDir)
		})

		It("doesn't restart the linuxkit process", func() {
			session := cf.Cf("dev", "start")
			Eventually(session, 10*time.Second).Should(gexec.Exit(0))

			Expect(session).To(gbytes.Say("CF Dev is already running..."))
			Expect(LaunchdPid("org.cloudfoundry.cfdev.linuxkit")).To(Equal(originalPid))
		})
	})
})

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
