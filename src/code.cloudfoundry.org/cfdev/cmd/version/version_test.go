package version_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = fmt.Sprintf(message, args...)
}

var _ = Describe("Version", func() {
	var (
		mockUI MockUI
		conf   config.Config
		verCmd *cobra.Command
	)

	BeforeEach(func() {
		mockUI = MockUI{WasCalledWith: ""}
		conf = config.Config{
			CliVersion: &semver.Version{Original: "1.2.3-rc.4"},
		}

		subject := &version.Version{
			UI:     &mockUI,
			Config: conf,
		}
		verCmd = subject.Cmd()
		verCmd.SetArgs([]string{})
	})

	It("prints the version", func() {
		Expect(verCmd.Execute()).To(Succeed())

		Expect(mockUI.WasCalledWith).To(Equal("Version: 1.2.3-rc.4"))
	})
})
