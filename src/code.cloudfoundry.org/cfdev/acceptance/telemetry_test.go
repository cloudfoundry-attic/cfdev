package acceptance

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"fmt"

	"path"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit starts and telemetry", func() {
	var (
		cfdevHome       string
		linuxkitPidPath string
		vpnkitPidPath   string
		stateDir        string
		cacheDir        string
	)

	BeforeEach(func() {
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

		fmt.Println(cfdevHome)
		os.RemoveAll(path.Join(cfdevHome, "analytics"))

		cmd := exec.Command("/usr/local/bin/cf", "install-plugin", os.Getenv("CFDEV_PLUGIN_PATH"), "-f")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	})

	AfterEach(func() {
		os.RemoveAll(path.Join(cfdevHome, "analytics"))
	})

	It("optout", func() {
		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "no")

		Eventually(session).Should(gbytes.Say("Downloading"))

		path := path.Join(cfdevHome, "analytics", "analytics.txt")
		contents, err := ioutil.ReadFile(path)

		Expect(string(contents)).Should(Equal("optout"))
		session.Kill()
	})

	It("optin", func() {
		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "yes")

		Eventually(session).Should(gbytes.Say("Downloading"))

		path := path.Join(cfdevHome, "analytics", "analytics.txt")
		contents, err := ioutil.ReadFile(path)

		Expect(string(contents)).Should(Equal("optin"))
		session.Kill()
	})

	It("is already opted in", func() {
		err := os.MkdirAll(path.Join(cfdevHome, "analytics"), 0755)
		Expect(err).ToNot(HaveOccurred())
		err = ioutil.WriteFile(path.Join(cfdevHome, "analytics", "analytics.txt"), []byte("optin"), 0755)
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command("/usr/local/bin/cf", "dev", "start")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)

		Expect(err).ToNot(HaveOccurred())
		Eventually(session).ShouldNot(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		Eventually(session).Should(gbytes.Say("Downloading"))

		path := path.Join(cfdevHome, "analytics", "analytics.txt")
		contents, err := ioutil.ReadFile(path)

		Expect(string(contents)).Should(Equal("optin"))
		session.Kill()
	})
})
