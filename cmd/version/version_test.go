package version_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cmd/start/mocks"
	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/iso"
	"code.cloudfoundry.org/cfdev/semver"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = fmt.Sprintf(message, args...)
}

var _ = Describe("Value", func() {
	var (
		mockUI         MockUI
		verCmd         *version.Version
		mockController *gomock.Controller
		mockIsoReader  *mocks.MockIsoReader
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIsoReader = mocks.NewMockIsoReader(mockController)
		mockUI = MockUI{WasCalledWith: ""}

		verCmd = &version.Version{
			UI:        &mockUI,
			IsoReader: mockIsoReader,
			Version:   &semver.Version{Original: "1.2.3-rc.4"},
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the cf-deps iso is not present", func() {
		It("prints the version", func() {
			verCmd.Execute(version.Args{DepsIsoPath: "/some-non-existent-file"})
			Expect(mockUI.WasCalledWith).To(Equal("CLI: 1.2.3-rc.4"))
		})
	})

	Context("when the cf-deps iso is present", func() {
		var (
			tmpFile string
		)

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "cfdev-version-test-")
			Expect(err).NotTo(HaveOccurred())
			tmpFile = f.Name()
		})

		AfterEach(func() {
			os.RemoveAll(tmpFile)
		})

		It("reports the versions in the metadata", func() {
			mockIsoReader.EXPECT().Read(tmpFile).Return(iso.Metadata{
				Versions: []iso.Version{
					{Name: "some-release-1", Value: "some-version-1"},
					{Name: "some-release-2", Value: "some-version-2"},
				},
			}, nil)

			verCmd.Execute(version.Args{DepsIsoPath: tmpFile})
			Expect(mockUI.WasCalledWith).To(ContainSubstring("CLI: 1.2.3-rc.4"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-1: some-version-1"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-2: some-version-2"))
		})
	})
})
