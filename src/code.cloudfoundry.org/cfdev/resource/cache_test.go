package resource_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"strings"

	"code.cloudfoundry.org/cfdev/resource"
)

type MockProgress struct {
	Total     uint64
	Current   uint64
	EndCalled bool
}

func (m *MockProgress) Write(b []byte) (int, error) { m.Current += uint64(len(b)); return len(b), nil }
func (m *MockProgress) Start(total uint64)          { m.Current = 0; m.Total = total }
func (m *MockProgress) Add(add uint64)              { m.Current += add }
func (m *MockProgress) End()                        { m.EndCalled = true }

var _ = Describe("Cache Sync", func() {

	var (
		tmpDir       string
		catalog      *resource.Catalog
		cache        *resource.Cache
		downloads    []string
		mockProgress *MockProgress
	)

	BeforeEach(func() {
		downloads = []string{}
		mockProgress = &MockProgress{}
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
					Size: 7,
				},
				{
					Name: "second-resource",
					URL:  "second-resource-url",
					MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
					Size: 7,
				},
				{
					Name: "third-resource",
					URL:  "third-resource-url",
					MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
					Size: 7,
				},
			},
		}

		createFile(tmpDir, "second-resource", "wrong-content")
		createFile(tmpDir, "third-resource", "content")

		cache = &resource.Cache{
			Dir: tmpDir,
			HttpDo: func(req *http.Request) (*http.Response, error) {
				downloads = append(downloads, req.URL.String())
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(strings.NewReader("content")),
				}, nil
			},
			Progress: mockProgress,
		}
	})

	It("downloads missing items to the target directory", func() {
		Expect(cache.Sync(catalog)).To(Succeed())

		Expect(downloads).To(ContainElement("first-resource-url"))
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "first-resource"))).To(Equal([]byte("content")))

		Expect(mockProgress.Total).To(Equal(uint64(21)))
	})

	It("re-downloads corrupt files to the target directory", func() {
		Expect(cache.Sync(catalog)).To(Succeed())

		Expect(downloads).To(ContainElement("second-resource-url"))
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "second-resource"))).To(Equal([]byte("content")))

	})

	It("does not re-download valid files and leaves file untouched", func() {
		Expect(cache.Sync(catalog)).To(Succeed())

		Expect(downloads).NotTo(ContainElement("third-resource-url"))
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "third-resource"))).To(Equal([]byte("content")))
	})

	It("informs progress", func() {
		Expect(cache.Sync(catalog)).To(Succeed())
		Expect(mockProgress.Total).To(Equal(uint64(21)))
		Expect(mockProgress.Current).To(Equal(uint64(21)))
	})

	It("resumes partially downloaded files to the target directory", func() {
		catalog.Items = catalog.Items[:1]
		createFile(tmpDir, "first-resource.tmp.9a0364b9e99bb480dd25e1f0284c8555", "cont")
		cache.HttpDo = func(req *http.Request) (*http.Response, error) {
			downloads = append(downloads, req.URL.String())
			Expect(req.Header).To(HaveKeyWithValue("Range", []string{"bytes=4-"}))
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(strings.NewReader("ent")),
			}, nil
		}

		Expect(cache.Sync(catalog)).To(Succeed())

		Expect(downloads).To(ContainElement("first-resource-url"))
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "first-resource"))).To(Equal([]byte("content")))
		Expect(filepath.Join(tmpDir, "first-resource.md5.9a0364b9e99bb480dd25e1f0284c8555")).ToNot(BeAnExistingFile())

		Expect(mockProgress.Total).To(Equal(uint64(7)))
		Expect(mockProgress.Current).To(Equal(uint64(7)))
	})

	It("handles file:// schema", func() {
		catalog = &resource.Catalog{Items: []resource.Item{{
			Name: "file-resource",
			URL:  fmt.Sprintf("file://%s/other-file", tmpDir),
			MD5:  "9a0364b9e99bb480dd25e1f0284c8555", // md5 -s content
			Size: 7,
		}}}
		Expect(ioutil.WriteFile(filepath.Join(tmpDir, "other-file"), []byte("content"), 0666)).To(Succeed())

		Expect(cache.Sync(catalog)).To(Succeed())

		Expect(downloads).To(BeEmpty())
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "file-resource"))).To(Equal([]byte("content")))

		Expect(mockProgress.Total).To(Equal(uint64(7)))
		Expect(mockProgress.Current).To(Equal(uint64(7)))
	})

	Context("when unknown resources are present", func() {
		BeforeEach(func() {
			createFile(tmpDir, "unknown-resource", "unknown-content")
		})
		It("deletes the unknown file", func() {
			Expect(cache.Sync(catalog)).To(Succeed())

			filename := filepath.Join(tmpDir, "unknown-resource")
			Expect(filename).ToNot(BeAnExistingFile())
		})
	})

	Context("cannot determine if a resources exists", func() {
		BeforeEach(func() {
			os.Chmod(tmpDir, 0222) // write only
		})
		It("returns an error", func() {
			err := cache.Sync(catalog)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("cannot determine checksum of a file", func() {
		BeforeEach(func() {
			os.Chmod(filepath.Join(tmpDir, "third-resource"), 0222)
		})
		It("returns an error", func() {
			err := cache.Sync(catalog)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("cannot delete corrupt file", func() {
		BeforeEach(func() {
			os.Chmod(tmpDir, 0400)
		})

		It("returns an error", func() {
			err := cache.Sync(catalog)
			Expect(err).To(HaveOccurred())
		})

		It("doesn't attempt the download", func() {
			Expect(downloads).ToNot(ContainElement("second-resource-url"))
		})
	})

	Context("downloading fails", func() {
		BeforeEach(func() {
			cache.HttpDo = func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Status:     "File Not Found",
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}, nil
			}
		})

		It("returns an error", func() {
			Expect(cache.Sync(catalog)).To(MatchError("http: File Not Found"))
		})
	})

	Context("downloaded file contains incorrect checksum", func() {
		BeforeEach(func() {
			cache.HttpDo = func(req *http.Request) (*http.Response, error) {
				downloads = append(downloads, req.URL.String())
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(strings.NewReader("wrong-content")),
				}, nil
			}
		})
		Context("file was originally missing", func() {
			BeforeEach(func() {
				catalog.Items = catalog.Items[:1]
			})
			It("returns an error", func() {
				err := cache.Sync(catalog)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("file was originally corrupt", func() {
			BeforeEach(func() {
				catalog.Items = catalog.Items[1:2]
			})
			It("returns an error", func() {
				err := cache.Sync(catalog)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("asset verification is turned off", func() {
		BeforeEach(func() {
			cache.SkipAssetVerification = true
		})

		It("does not delete files with different checksums", func() {
			Expect(cache.Sync(catalog)).To(Succeed())

			corruptFile := filepath.Join(tmpDir, "second-resource")
			Expect(ioutil.ReadFile(corruptFile)).To(Equal([]byte("wrong-content")))
		})

		It("doesn't re-download files with different checksums", func() {
			Expect(cache.Sync(catalog)).To(Succeed())
			Expect(downloads).ToNot(ContainElement("second-resource-url"))
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
