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

		catalog = &resource.Catalog{
			Items: []resource.Item{
				{
					Name: "first-resource",
					URL:  "first-resource-url",
					MD5:  "60484def35e5e27a5bda4f88dd5434d3", // md5 -s first-resource
				},
				{
					Name: "second-resource",
					URL:  "second-resource-url",
					MD5:  "some-corrupt-md5",
				},
				{
					Name: "third-resource",
					URL:  "third-resource-url",
					MD5:  "db0495de5b3a1d9eb92bfdc7bbe38564", // md5 -s third-resource
				},
			},
		}

		createFile(tmpDir, "second-resource")
		createFile(tmpDir, "third-resource")

		cache = &resource.Cache{
			Dir: tmpDir,
			DownloadFunc: func(url, path string) error {
				downloads = append(downloads, download{url, path})
				return nil
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

	It("deletes corrupt files", func() {
		corruptFile := filepath.Join(tmpDir, "second-resource")
		Expect(corruptFile).ToNot(BeAnExistingFile())
	})

	It("re-downloads corrupt files to the target directory", func() {
		Expect(downloads).To(ContainElement(download{
			url:  "second-resource-url",
			path: filepath.Join(tmpDir, "second-resource"),
		}))
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
			createFile(tmpDir, "unknown-resource")
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
})

func createFile(path string, name string) {
	filename := filepath.Join(path, name)
	err := ioutil.WriteFile(filename, []byte(name), 0777)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
}

type download struct {
	url  string
	path string
}
