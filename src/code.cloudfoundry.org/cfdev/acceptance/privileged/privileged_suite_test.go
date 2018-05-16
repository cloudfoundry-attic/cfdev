package privileged_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"

	"github.com/onsi/gomega/gexec"
)

var pluginPath string
var cfdevHome string
var hyperkitPidPath string

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance - privileged suite")
}

var _ = BeforeSuite(func() {
	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		Fail("please provide CFDEV_PLUGIN_PATH (use ./generate-plugin.sh)")
	}
	os.Unsetenv("BOSH_ALL_PROXY")

	SetDefaultEventuallyTimeout(10 * time.Second)

	Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")
	RemoveIPAliases(BoshDirectorIP, CFRouterIP)

	cfdevHome = os.Getenv("CFDEV_HOME")
	if cfdevHome == "" {
		cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
	}
	hyperkitPidPath = filepath.Join(cfdevHome, "state", "linuxkit", "hyperkit.pid")

	session := cf.Cf("install-plugin", pluginPath, "-f")
	Eventually(session).Should(gexec.Exit(0))
})

var _ = AfterSuite(func() {
	session := cf.Cf("uninstall-plugin", "cfdev")
	Eventually(session).Should(gexec.Exit(0))
	gexec.CleanupBuildArtifacts()
})

func HasSudoPrivilege() bool {
	cmd := exec.Command("sh", "-c", "sudo -n true")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	if session.ExitCode() == 0 {
		return true
	}
	return false
}
