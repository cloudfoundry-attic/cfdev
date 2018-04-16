package privileged_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"io/ioutil"
	"syscall"

	"time"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("hyperkit lifecycle", func() {
	var (
		cfdevHome       string
		linuxkitPidPath string
		vpnkitPidPath   string
		stateDir        string
		cacheDir        string
	)

	BeforeEach(func() {
		Expect(HasSudoPrivilege()).To(BeTrue())
		RemoveIPAliases(BoshDirectorIP, CFRouterIP)

		cfHome, err := ioutil.TempDir("", "cf-home")
		Expect(err).ToNot(HaveOccurred())

		cfdevHome = CreateTempCFDevHomeDir()
		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")
		linuxkitPidPath = filepath.Join(stateDir, "linuxkit.pid")
		vpnkitPidPath = filepath.Join(stateDir, "vpnkit.pid")

		if os.Getenv("CFDEV_PLUGIN_PATH") == "" {
			SetupDependencies(cacheDir)
			os.Setenv("CFDEV_SKIP_ASSET_CHECK", "true")
		}
		os.Setenv("CF_HOME", cfHome)
		os.Setenv("CFDEV_HOME", cfdevHome)

		session := cf.Cf("install-plugin", pluginPath, "-f")
		Eventually(session).Should(gexec.Exit(0))
		session = cf.Cf("plugins")
		Eventually(session).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		gexec.KillAndWait()
		vmPid := PidFromFile(linuxkitPidPath)
		vpnPid := PidFromFile(vpnkitPidPath)

		if vmPid != 0 {
			syscall.Kill(int(-vmPid), syscall.SIGKILL)
		}

		if vpnPid != 0 {
			syscall.Kill(int(-vpnPid), syscall.SIGKILL)
		}

		os.RemoveAll(cfdevHome)
		RemoveIPAliases(BoshDirectorIP, CFRouterIP)

		session := cf.Cf("uninstall-plugin", "cfdev")
		Eventually(session).Should(gexec.Exit(0))
	})

	It("runs the entire vm lifecycle", func() {
		session := cf.Cf("dev", "start")
		Eventually(session, 20*time.Minute).Should(gbytes.Say("Starting VPNKit"))

		By("settingup VPNKit dependencies")
		Eventually(filepath.Join(cfdevHome, "http_proxy.json"), 10, 1).Should(BeAnExistingFile())

		Eventually(vpnkitPidPath, 10, 1).Should(BeAnExistingFile())
		Eventually(linuxkitPidPath, 10, 1).Should(BeAnExistingFile())

		// FYI - this will take time until we use thin provisioned disks
		hyperkitPidPath := filepath.Join(stateDir, "hyperkit.pid")
		Eventually(hyperkitPidPath, 120, 1).Should(BeAnExistingFile())

		By("waiting for garden to listen")
		EventuallyShouldListenAt("http://"+GardenIP+":8888", 240)

		EventuallyWeCanTargetTheBOSHDirector()

		By("waiting for cf router to listen")
		EventuallyShouldListenAt("http://"+CFRouterIP+":80", 1200)

		By("waiting for cfdev cli to exit when the deploy finished")
		Eventually(session, 300).Should(gexec.Exit(0))

		linuxkitPid := PidFromFile(linuxkitPidPath)
		hyperkitPid := PidFromFile(hyperkitPidPath)
		vpnkitPid := PidFromFile(vpnkitPidPath)

		By("deploy finished - stopping...")
		session = cf.Cf("dev", "stop")
		Eventually(session).Should(gexec.Exit(0))

		//ensure pid is not running
		EventuallyProcessStops(linuxkitPid, 5)
		EventuallyProcessStops(hyperkitPid, 5)
		EventuallyProcessStops(vpnkitPid, 5)
	})
})

func EventuallyWeCanTargetTheBOSHDirector() {
	By("waiting for bosh to listen")
	EventuallyShouldListenAt("https://"+BoshDirectorIP+":25555", 480)

	// Even though the test below is very similar this fails fast when `bosh env`
	// command is broken

	session := cf.Cf("dev", "bosh", "env")
	Eventually(session).Should(gexec.Exit(0))

	// This test is more representative of how `bosh env` should be invoked
	w := gexec.NewPrefixedWriter("[bosh env] ", GinkgoWriter)
	boshCmd := exec.Command("/bin/sh",
		"-e",
		"-c", fmt.Sprintf(`eval "$(cf dev bosh env)" && bosh env`))

	session, err := gexec.Start(boshCmd, w, w)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 30, 1).Should(gexec.Exit(0))
}

func RemoveIPAliases(aliases ...string) {
	for _, alias := range aliases {
		cmd := exec.Command("sudo", "-n", "ifconfig", "lo0", "inet", alias+"/32", "remove")
		writer := gexec.NewPrefixedWriter("[ifconfig] ", GinkgoWriter)
		session, err := gexec.Start(cmd, writer, writer)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}
