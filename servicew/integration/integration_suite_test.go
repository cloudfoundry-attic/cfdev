package integration_test

import (
	"github.com/onsi/gomega/gexec"
	"io"
	"os"
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
	binaryPath, err = gexec.Build("code.cloudfoundry.org/cfdev/servicew")
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

func copy(src string, dest string, chmod ...bool) error {
	target, err := os.Create(dest)
	Expect(err).NotTo(HaveOccurred())
	defer target.Close()

	if len(chmod) != 0 && chmod[0] {
		err = os.Chmod(dest, 0744)
		Expect(err).NotTo(HaveOccurred())
	}

	binData, err := os.Open(src)
	Expect(err).NotTo(HaveOccurred())
	defer binData.Close()

	_, err = io.Copy(target, binData)
	return err
}
