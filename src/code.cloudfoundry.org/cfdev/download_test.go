package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cfdev/resource"
)

var _ = Describe("download acceptance", func() {
	var (
		cfdevHome string
		cacheDir  string
		server    *httptest.Server
	)

	BeforeEach(func() {
		cfdevHome = createTempCFDevHomeDir()
		cacheDir = filepath.Join(cfdevHome, "cache")

		serverAssetsDir := stageServerAssets()
		fileHandler := http.FileServer(http.Dir(serverAssetsDir))
		server = httptest.NewServer(fileHandler)
	})

	AfterEach(func() {
		gexec.KillAndWait()
		os.RemoveAll(cfdevHome)
		server.Close()
	})

	It("downloads assets", func() {
		command := exec.Command(cliPath, "download")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome),
			fmt.Sprintf("CFDEV_CATALOG=%s", localCatalog(server.URL)),
		)

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, 10, 1).Should(gexec.Exit(0))

		files, err := ioutil.ReadDir(cacheDir)
		Expect(err).ToNot(HaveOccurred())

		Expect(names(files)).To(ConsistOf("some-asset"))
	})

	Context("downloaded asset has incorrect checksum", func() {
		It("should exit", func() {
			command := exec.Command(cliPath, "download")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome),
				fmt.Sprintf("CFDEV_CATALOG=%s", badCatalog(server.URL)),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, 10, 1).Should(gexec.Exit(1))

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
				URL:  fmt.Sprintf("%s/%s", serverAddr, "some-asset"),
				Name: "some-asset",
				MD5:  "incorrect-md5",
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
				URL:  fmt.Sprintf("%s/%s", serverAddr, "some-asset"),
				Name: "some-asset",
				MD5:  "ad60407c083b4ecc372614b8fcd9f305",
			},
		},
	}

	bytes, err := json.Marshal(c)
	Expect(err).ToNot(HaveOccurred())

	return string(bytes)
}
