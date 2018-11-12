package version_test

import (
	"fmt"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cmd/version/mocks"
	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/metadata"
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
		mockUI             MockUI
		verCmd             *version.Version
		mockController     *gomock.Controller
		mockMetaDataReader *mocks.MockMetaDataReader
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockMetaDataReader = mocks.NewMockMetaDataReader(mockController)
		mockUI = MockUI{WasCalledWith: ""}

		verCmd = &version.Version{
			UI:             &mockUI,
			MetaDataReader: mockMetaDataReader,
			Version:        &semver.Version{Original: "1.2.3-rc.4"},
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Context("when the metadata file is not present", func() {
		It("prints the version", func() {
			verCmd.Execute()
			Expect(mockUI.WasCalledWith).To(Equal("CLI: 1.2.3-rc.4"))
		})
	})

	Context("when the metadata file is present", func() {
		var (
			cacheDir string
		)

		BeforeEach(func() {
			var err error
			cacheDir, err = ioutil.TempDir("", "cfdev-version-cache-test-")
			Expect(err).NotTo(HaveOccurred())

			f, err := os.Create(filepath.Join(cacheDir, "metadata.yml"))
			Expect(err).NotTo(HaveOccurred())
			f.Close()

			verCmd.Config.CacheDir = cacheDir
		})

		AfterEach(func() {
			os.RemoveAll(cacheDir)
		})

		It("reports the versions in the metadata", func() {
			mockMetaDataReader.EXPECT().Read(filepath.Join(cacheDir, "metadata.yml")).Return(metadata.Metadata{
				Versions: []metadata.Version{
					{Name: "some-release-1", Value: "some-version-1"},
					{Name: "some-release-2", Value: "some-version-2"},
				},
			}, nil)

			verCmd.Execute()
			Expect(mockUI.WasCalledWith).To(ContainSubstring("CLI: 1.2.3-rc.4"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-1: some-version-1"))
			Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-2: some-version-2"))
		})
	})
})
