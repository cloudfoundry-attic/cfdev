package acceptance

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var (
	pluginPath                string
	tmpDir, cfHome, cfdevHome string
)

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance suite")
}

var _ = BeforeSuite(func() {
	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		Fail("Please set CFDEV_PLUGIN_PATH env var to a fully qualified path to a valid plugin")
	}
	SetDefaultEventuallyTimeout(5 * time.Second)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
	Expect(os.RemoveAll(tmpDir)).To(Succeed())
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

	session := cf.Cf("install-plugin", pluginPath, "-f")
	Eventually(session).Should(gexec.Exit(0))
	session = cf.Cf("plugins")
	Eventually(session).Should(gbytes.Say("cfdev"))
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterEach(func() {
	session := cf.Cf("dev", "stop")
	Eventually(session, 30, 1).Should(gexec.Exit())
	os.Unsetenv("CF_HOME")
	os.Unsetenv("CFDEV_HOME")
})
