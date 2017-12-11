package resource_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cli/cf/errors"
)

var _ = Describe("Cache Sync", func() {

	var (
		tmpDir  string
		catalog *resource.Catalog
		cache   *resource.Cache

		err       error
		downloads []download
	)

	BeforeEach(func() {
		downloads = nil
		tmpDir, _ = ioutil.TempDir("", "scan")

		// This catalog is representative of the different actions
		// the cache will encounter
		// 1. Asset is missing
		// 2. Existing asset contains incorrect checksum
		// 3. Existing asset contains correct checksum
		catalog = &resource.Catalog{
			Items: []resource.Item{
				{
					Name: "first-resource",
					URL:  "first-resource-url",
					MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
				},
				{
					Name: "second-resource",
					URL:  "second-resource-url",
					MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
				},
				{
					Name: "third-resource",
					URL:  "third-resource-url",
					MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
				},
			},
		}

		createFile(tmpDir, "second-resource", "wrong-content")
		createFile(tmpDir, "third-resource", "content")

		cache = &resource.Cache{
			Dir: tmpDir,
			DownloadFunc: func(url, path string) error {
				downloads = append(downloads, download{url, path})
				return ioutil.WriteFile(path, []byte("content"), 0777)
			},
		}
	})

	JustBeforeEach(func() {
		err = cache.Sync(catalog)
	})

	It("downloads missing items to the target directory", func() {
		Expect(downloads).To(ContainElement(download{
			url:  "first-resource-url",
			path: filepath.Join(tmpDir, "first-resource"),
		}))
	})

	It("re-downloads corrupt files to the target directory", func() {
		originallyCorrupt := filepath.Join(tmpDir, "second-resource")

		Expect(downloads).To(ContainElement(download{
			url:  "second-resource-url",
			path: originallyCorrupt,
		}))

		Expect(ioutil.ReadFile(originallyCorrupt)).To(Equal([]byte("content")))
	})

	It("leaves valid files untouched", func() {
		validFile := filepath.Join(tmpDir, "third-resource")
		Expect(validFile).To(BeAnExistingFile())
	})

	It("does not re-download valid files", func() {
		Expect(downloads).ToNot(ContainElement(download{
			url:  "third-resource-url",
			path: filepath.Join(tmpDir, "third-resource"),
		}))
	})

	Context("when unknown resources are present", func() {
		BeforeEach(func() {
			createFile(tmpDir, "unknown-resource", "unknown-content")
		})

		It("deletes the unknown file", func() {
			filename := filepath.Join(tmpDir, "unknown-resource")
			Expect(filename).ToNot(BeAnExistingFile())

		})
	})

	Context("cannot determine if a resources exists", func() {
		BeforeEach(func() {
			os.Chmod(tmpDir, 0222) // write only
		})

		AfterEach(func() {
			os.Chmod(tmpDir, 0777)
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("cannot determine checksum of a file", func() {
		BeforeEach(func() {
			resource := filepath.Join(tmpDir, "third-resource")
			os.Chmod(resource, 0222)
		})

		AfterEach(func() {
			resource := filepath.Join(tmpDir, "third-resource")
			os.Chmod(resource, 0777)
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("cannot delete corrupt file", func() {
		BeforeEach(func() {
			os.Chmod(tmpDir, 0400)
		})

		AfterEach(func() {
			os.Chmod(tmpDir, 0777)
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("doesn't attempt the download", func() {
			Expect(downloads).ToNot(ContainElement(download{
				url:  "second-resource-url",
				path: filepath.Join(tmpDir, "second-resource"),
			}))
		})
	})

	Context("downloading fails", func() {
		BeforeEach(func() {
			cache.DownloadFunc = func(url, path string) error {
				return errors.New("unable to download")
			}
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("unable to download"))
		})
	})

	Context("downloaded file contains incorrect checksum", func() {
		Context("file was originally missing", func() {
			BeforeEach(func() {
				cache.DownloadFunc = func(url, path string) error {
					createFile(tmpDir, "first-resource", "wrong-content")
					createFile(tmpDir, "second-resource", "second-content")
					return nil
				}
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("file was originally corrupt", func() {
			BeforeEach(func() {
				cache.DownloadFunc = func(url, path string) error {
					createFile(tmpDir, "first-resource", "first-content")
					createFile(tmpDir, "second-resource", "wrong-content")
					return nil
				}

			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("asset verification is turned off", func() {
		BeforeEach(func() {
			cache.SkipAssetVerification = true
		})

		It("does not delete files with different checksums", func() {
			corruptFile := filepath.Join(tmpDir, "second-resource")
			Expect(corruptFile).To(BeAnExistingFile())
		})

		It("doesn't re-download files with different checksums", func() {
			Expect(downloads).ToNot(ContainElement(download{
				url:  "second-resource-url",
				path: filepath.Join(tmpDir, "second-resource"),
			}))
		})

	})
})

func createFile(dir, name, contents string) {
	filename := filepath.Join(dir, name)
	err := ioutil.WriteFile(filename, []byte(contents), 0777)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

type download struct {
	url  string
	path string
}
