package privileged_test

import (
	. "code.cloudfoundry.org/cfdev/acceptance"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"testing"
	"time"
)

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance - privileged suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)

	Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")

	pluginPath := os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		Fail("Please set CFDEV_PLUGIN_PATH env var to a fully qualified path to a valid plugin")
	}

	session := cf.Cf("install-plugin", "-f", pluginPath)
	<-session.Exited

	os.Unsetenv("BOSH_ALL_PROXY")
})

var _ = AfterSuite(func() {
	cf.Cf("uninstall-plugin", "cfdev")
})