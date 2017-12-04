package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

func TestCFDev(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cfdev suite")
}

var (
	cliPath string
)

var _ = BeforeSuite(func() {
	var err error
	cliPath, err = gexec.Build("code.cloudfoundry.org/cfdev")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
