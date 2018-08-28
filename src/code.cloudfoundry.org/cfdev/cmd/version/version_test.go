package version_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cmd/start/mocks"
	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/iso"
	"code.cloudfoundry.org/cfdev/semver"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
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
			UI:      &mockUI,
			Config:  config.Config{CacheDir: "/some-non-existent-dir"},
			IsoReader: mockIsoReader,
			Version: &semver.Version{Original: "1.2.3-rc.4"},
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	It("prints the version", func() {
		verCmd.Execute()
		Expect(mockUI.WasCalledWith).To(Equal("Value: 1.2.3-rc.4"))
	})

	Context("when the cf-deps iso is present in cache dir", func() {
		var (
			tmpDir string
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "cfdev-version-test-")
			Expect(err).NotTo(HaveOccurred())

			verCmd.Config.CacheDir = tmpDir
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		FIt("reports the versions in the metadata", func() {
			_, err := os.Create(filepath.Join(tmpDir, "cf-deps.iso"))
			Expect(err).NotTo(HaveOccurred())

			mockIsoReader.EXPECT().Read(filepath.Join(tmpDir, "cf-deps.iso")).Return(iso.Metadata{
				Versions: []iso.Version{
					{Name: "some-release-1", Value: "some-version-1"},
					{Name: "some-release-2", Value: "some-version-2"},
				},
			}, nil)

			verCmd.Execute()
			Expect(mockUI.WasCalledWith).To(ContainSubstring("Value: 1.2.3-rc.4"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-1: some-version-1"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-2: some-version-2"))
		})
	})
})
