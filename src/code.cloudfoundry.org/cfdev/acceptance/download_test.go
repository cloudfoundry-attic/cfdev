package acceptance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cfdev/resource"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
)

var _ = Describe("download", func() {
	var (
		cacheDir string
		server   *httptest.Server
	)

	BeforeEach(func() {
		cacheDir = filepath.Join(cfdevHome, "cache")

		serverAssetsDir := stageServerAssets()
		fileHandler := http.FileServer(http.Dir(serverAssetsDir))
		server = httptest.NewServer(fileHandler)
	})

	AfterEach(func() {
		gexec.Kill()
		server.Close()
	})

	Context("when the catalog is valid", func() {
		BeforeEach(func() { os.Setenv("CFDEV_CATALOG", localCatalog(server.URL)) })
		AfterEach(func() { os.Unsetenv("CFDEV_CATALOG") })

		It("downloads assets", func() {
			session := cf.Cf("dev", "download")
			Eventually(session, 10, 1).Should(gexec.Exit(0))

			files, err := ioutil.ReadDir(cacheDir)
			Expect(err).ToNot(HaveOccurred())

			Expect(names(files)).To(ConsistOf("some-asset"))
		})
	})

	Context("downloaded asset has incorrect checksum", func() {
		BeforeEach(func() { os.Setenv("CFDEV_CATALOG", badCatalog(server.URL)) })
		AfterEach(func() { os.Unsetenv("CFDEV_CATALOG") })

		It("should exit", func() {
			session := cf.Cf("dev", "download")
			Eventually(session, 10).Should(gexec.Exit(1))
		})
	})
})

func names(infos []os.FileInfo) []string {
	names := make([]string, 0, len(infos))

	for _, info := range infos {
		names = append(names, info.Name())
	}

	return names
}

func stageServerAssets() string {
	dir, err := ioutil.TempDir("", "cfdev-server-assets")
	Expect(err).ToNot(HaveOccurred())

	filename := filepath.Join(dir, "some-asset")
	err = ioutil.WriteFile(filename, []byte("some-content"), 0777)
	Expect(err).ToNot(HaveOccurred())

	return dir
}

func badCatalog(serverAddr string) string {
	c := &resource.Catalog{
		Items: []resource.Item{
			{
				URL:   fmt.Sprintf("%s/%s", serverAddr, "some-asset"),
				Name:  "some-asset",
				MD5:   "incorrect-md5",
				Size:  uint64(len("some-content")),
				InUse: true,
			},
		},
	}

	bytes, err := json.Marshal(c)
	Expect(err).ToNot(HaveOccurred())

	return string(bytes)
}

func localCatalog(serverAddr string) string {
	c := &resource.Catalog{
		Items: []resource.Item{
			{
				URL:   fmt.Sprintf("%s/%s", serverAddr, "some-asset"),
				Name:  "some-asset",
				MD5:   "ad60407c083b4ecc372614b8fcd9f305",
				Size:  uint64(len("some-content")),
				InUse: true,
			},
		},
	}

	bytes, err := json.Marshal(c)
	Expect(err).ToNot(HaveOccurred())

	return string(bytes)
}
