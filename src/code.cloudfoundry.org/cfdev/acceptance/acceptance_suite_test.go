package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var pluginPath string

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance suite")
}

var _ = BeforeSuite(func() {
	var err error
	pluginPath, err = gexec.Build("code.cloudfoundry.org/cfdev")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
