package acceptance

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/onsi/gomega/gexec"
)

var (
	pluginPath                string
	tmpDir, cfHome, cfdevHome string
	cfPluginHome              string
)

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(time.Minute)

	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		Fail("Please set CFDEV_PLUGIN_PATH env var to a fully qualified path to a valid plugin")
	}

	var err error
	cfPluginHome, err = ioutil.TempDir("", "cfdev-plugin-")
	Expect(err).ToNot(HaveOccurred())
	os.Setenv("CF_PLUGIN_HOME", cfPluginHome)

	session := cf.Cf("install-plugin", pluginPath, "-f")
	Eventually(session).Should(gexec.Exit())
})

var _ = AfterSuite(func() {
	cf.Cf("uninstall-plugin", "cfdev")
	os.Unsetenv("CF_PLUGIN_HOME")
})

var _ = BeforeEach(func() {
	var err error
	tmpDir, err = ioutil.TempDir("", "cfdev-acceptance-")
	Expect(err).ToNot(HaveOccurred())

	cfHome = filepath.Join(tmpDir, "cf_home")
	Expect(os.Mkdir(cfHome, 0755)).To(Succeed())
	os.Setenv("CF_HOME", cfHome)

	cfdevHome = filepath.Join(tmpDir, "cfdev_home")
	Expect(os.Mkdir(cfdevHome, 0755)).To(Succeed())
	os.Setenv("CFDEV_HOME", cfdevHome)
})

var _ = AfterEach(func() {
	session := cf.Cf("dev", "stop")
	Eventually(session).Should(gexec.Exit())

	if IsWindows() {
		exec.Command("powershell.exe", "-Command", "Stop-Process -Name cfdev,cf -Force -EA 0").Run()
	}

	Expect(os.RemoveAll(tmpDir)).To(Succeed())
	os.Unsetenv("CF_HOME")
	os.Unsetenv("CFDEV_HOME")
})
