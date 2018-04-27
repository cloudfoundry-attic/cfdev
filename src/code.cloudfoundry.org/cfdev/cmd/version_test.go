package cmd_test

import (
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

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

		verCmd = cmd.NewVersion(&mockUI, conf)
		verCmd.SetArgs([]string{})
	})

	It("prints the version", func() {
		Expect(verCmd.Execute()).To(Succeed())

		Expect(mockUI.WasCalledWith).To(Equal("Version: 1.2.3-rc.4"))
	})
})
