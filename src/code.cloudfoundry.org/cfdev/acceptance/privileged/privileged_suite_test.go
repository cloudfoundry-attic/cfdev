package privileged_test

import (
	"testing"
	"time"

	. "code.cloudfoundry.org/cfdev/acceptance"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os"
)

var pluginPath string
var cfdevHome string
var hyperkitPidPath string

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance - privileged suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")
	RemoveIPAliases(BoshDirectorIP, CFRouterIP)

	var err error

	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		Fail("Please set CFDEV_PLUGIN_PATH env var to a fully qualified path to a valid plugin")
	}

	os.Setenv("CFDEV_PLUGIN_PATH", pluginPath)

	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})



