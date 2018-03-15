package privileged_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var pluginPath string

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
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
