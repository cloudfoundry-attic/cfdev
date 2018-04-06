package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"time"

	"github.com/onsi/gomega/gexec"
)

var pluginPath string

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance suite")
}

var _ = BeforeSuite(func() {
	pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
	if pluginPath == "" {
		var err error
		pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev")
		Expect(err).ShouldNot(HaveOccurred())
	}

	SetDefaultEventuallyTimeout(5 * time.Second)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
