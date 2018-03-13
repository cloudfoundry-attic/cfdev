package resource

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Start", func() {

	var (
		downloader Downloader
		server     *ghttp.Server
		err        error
		targetPath string
		url        string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AllowUnhandledRequests = true

		tmpDir, _ := ioutil.TempDir("", "file-downloader")
		targetPath = filepath.Join(tmpDir, "target-resource-file")

		url = server.URL() + "/resource"
	})

	AfterEach(func() {
		server.Close()
	})

	JustBeforeEach(func() {
		err = downloader.Start(url, targetPath)
	})

	It("attempts to download the file", func() {
		Expect(server.ReceivedRequests()).To(HaveLen(1))
		req := server.ReceivedRequests()[0]
		Expect(req.Method).To(Equal("GET"))
		Expect(req.URL.Path).To(Equal("/resource"))
	})

	Context("unable to connect to the resource server", func() {
		BeforeEach(func() {
			url = "http://some-bad-address"
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("resource server returns does not return HTTP OK", func() {
		BeforeEach(func() {
			server.AppendHandlers(ghttp.RespondWith(http.StatusBadRequest, nil))
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when downloading succeeds", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.RespondWith(http.StatusOK, "some-content"),
			)
		})

		It("saves the content to the target path", func() {
			Expect(targetPath).To(BeARegularFile())

			content, err := ioutil.ReadFile(targetPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal([]byte("some-content")))
		})

		Context("when target path doesn't have write permissions", func() {
			BeforeEach(func() {
				parentFolder := filepath.Dir(targetPath)
				os.Chmod(parentFolder, 0555) // read+execute
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&os.PathError{}))
			})
		})
	})

	Context("when the connection drops while transferring", func() {
		BeforeEach(func() {
			server.AppendHandlers(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)

				flusher := w.(http.Flusher)
				flusher.Flush()

				server.CloseClientConnections()
			})
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
		})

		It("removes the partially downloaded file", func() {
			Expect(targetPath).ToNot(BeARegularFile())
		})
	})

	Context("url is a local file path", func() {
		var testDir string
		BeforeEach(func() {
			testDir, err = ioutil.TempDir("", "cfdev.downloader.")
			Expect(err).ToNot(HaveOccurred())
			Expect(ioutil.WriteFile(filepath.Join(testDir, "file"), []byte("contents"), 0644)).To(Succeed())

			contents, err := ioutil.ReadFile(filepath.Join(testDir, "file"))
			Expect(err).ToNot(HaveOccurred())

			Expect(contents).To(Equal([]byte("contents")))
			url = filepath.Join("file:/", testDir, "file")
		})
		AfterEach(func() { os.RemoveAll(testDir) })

		It("should copy local file", func() {
			Expect(ioutil.ReadFile(filepath.Join(testDir, "file"))).To(Equal([]byte("contents")))
			Expect(ioutil.ReadFile(targetPath)).To(Equal([]byte("contents")))
		})
	})
})
