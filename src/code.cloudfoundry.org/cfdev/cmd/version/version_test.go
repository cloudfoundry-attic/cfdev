package version_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		verCmd *version.Version
	)

	BeforeEach(func() {
		mockUI = MockUI{WasCalledWith: ""}
		verCmd = &version.Version{
			UI:      &mockUI,
			Version: &semver.Version{Original: "1.2.3-rc.4"},
		}
	})

	It("prints the version", func() {
		verCmd.Execute()
		Expect(mockUI.WasCalledWith).To(Equal("Version: 1.2.3-rc.4"))
	})
})
