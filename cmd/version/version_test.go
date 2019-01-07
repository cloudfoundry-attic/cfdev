package version_test

import (
	"code.cloudfoundry.org/cfdev/resource"
	"fmt"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/cmd/version/mocks"
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
	if m.WasCalledWith == "" {
		m.WasCalledWith = fmt.Sprintf(message, args...)
		return
	}
	m.WasCalledWith = m.WasCalledWith + "\n" + fmt.Sprintf(message, args...)
}

var _ = Describe("Version Command", func() {
	var (
		mockUI             MockUI
		verCmd             *version.Version
		mockController     *gomock.Controller
		mockMetaDataReader *mocks.MockMetaDataReader
	)
	Context("when no -f flag is passed", func() {

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockMetaDataReader = mocks.NewMockMetaDataReader(mockController)
			mockUI = MockUI{WasCalledWith: ""}

			verCmd = &version.Version{
				UI:             &mockUI,
				MetaDataReader: mockMetaDataReader,
				Version:        &semver.Version{Original: "1.2.3-rc.4"},
				BuildVersion:   "some-build-version",
			}
		})

		AfterEach(func() {
			mockController.Finish()
		})

		Context("when the metadata file is not present", func() {
			It("prints only the version of CLI", func() {
				verCmd.Execute("")
				Expect(mockUI.WasCalledWith).To(ContainSubstring("CLI: 1.2.3-rc.4"))
				Expect(mockUI.WasCalledWith).To(ContainSubstring("BUILD: some-build-version"))
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

				verCmd.Execute("")
				Expect(mockUI.WasCalledWith).To(ContainSubstring("CLI: 1.2.3-rc.4"))
				Expect(mockUI.WasCalledWith).To(ContainSubstring("BUILD: some-build-version"))
				Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-1: some-version-1"))
				Expect(mockUI.WasCalledWith).To(ContainSubstring("some-release-2: some-version-2"))
			})
		})
	})

	Context("when a -f flag is passed", func() {
		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockMetaDataReader = mocks.NewMockMetaDataReader(mockController)
			mockUI = MockUI{WasCalledWith: ""}

			verCmd = &version.Version{
				UI:             &mockUI,
				MetaDataReader: metadata.New(),
				Version:        &semver.Version{Original: "1.2.3-rc.4"},
			}
		})

		AfterEach(func() {
			mockController.Finish()
		})

		Context("when there is no file at the path provided", func() {
			It("returns an error", func() {
				verCmd.Execute("/some-bad-filepath")
				Expect(mockUI.WasCalledWith).To(ContainSubstring("file not found"))
			})
		})

		Context("when there is a file at the path provided", func() {
			Context("when the tarball contains valid meta data", func() {
				var (
					tarFilepath string
					folderToTar string
					targetTar   string
				)

				BeforeEach(func() {
					var err error

					folderToTar, err = ioutil.TempDir("", "folderToTar")
					Expect(err).ToNot(HaveOccurred())
					Expect(os.Mkdir(filepath.Join(folderToTar, "deployment_config"), 0766)).ToNot(HaveOccurred())

					targetTar, err = ioutil.TempDir("", "targetTar")
					Expect(err).ToNot(HaveOccurred())

					yml := `versions:
  - name: some-name
    version: some-version`
					err = ioutil.WriteFile(filepath.Join(folderToTar, "deployment_config", "metadata.yml"), []byte(yml), 0666)
					Expect(err).ToNot(HaveOccurred())

					tarFilepath = filepath.Join(targetTar, "deps.tgz")
					tarDst, err := os.Create(tarFilepath)
					Expect(err).NotTo(HaveOccurred())
					defer tarDst.Close()

					Expect(resource.Tar(folderToTar, tarDst)).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.RemoveAll(folderToTar)
					os.RemoveAll(targetTar)
				})

				It("returns an error", func() {
					verCmd.Execute(tarFilepath)
					Expect(mockUI.WasCalledWith).To(ContainSubstring("some-name: some-version"))
				})
			})

			Context("when the tarball does not contain meta data", func() {
				var (
					tarFilepath string
					folderToTar string
					targetTar   string
				)

				BeforeEach(func() {
					var err error

					folderToTar, err = ioutil.TempDir("", "otherFolderToTar")
					Expect(err).ToNot(HaveOccurred())

					targetTar, err = ioutil.TempDir("", "otherTargetTar")
					Expect(err).ToNot(HaveOccurred())

					tarFilepath = filepath.Join(targetTar, "deps.tgz")
					tarDst, err := os.Create(tarFilepath)
					Expect(err).NotTo(HaveOccurred())
					defer tarDst.Close()

					Expect(resource.Tar(folderToTar, tarDst)).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					os.RemoveAll(folderToTar)
					os.RemoveAll(targetTar)
				})

				It("returns an error", func() {
					verCmd.Execute(tarFilepath)
					Expect(mockUI.WasCalledWith).To(ContainSubstring("Metadata not found version unknown"))
				})
			})
		})
	})
})
