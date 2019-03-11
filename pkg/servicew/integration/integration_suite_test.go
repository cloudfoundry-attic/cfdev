package integration_test

import (
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var binaryPath string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceWrapper Integration Suite")
}

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("code.cloudfoundry.org/cfdev/pkg/servicew")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func run(executable string, args ...string) string {
	command := exec.Command(executable, args...)
	output, err := command.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
	return string(output)
}

func fixturePath(name string) string {
	return filepath.Join("fixtures", name)
}