package main_test

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

		fileHandler := http.FileServer(http.Dir(testResourcePath()))
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
		Eventually(session, 1200, 1).Should(gexec.Exit(0))

		files, err := ioutil.ReadDir(cacheDir)
		Expect(err).ToNot(HaveOccurred())

		Expect(names(files)).To(ConsistOf("bosh-deps.iso", "cf-deps.iso", "cfdev-efi.iso"))
	})
})

func names(infos []os.FileInfo) []string {
	names := make([]string, 0, len(infos))

	for _, info := range infos {
		names = append(names, info.Name())
	}

	return names
}

func testResourcePath() string {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	return filepath.Join(gopaths[0], "linuxkit")
}

func localCatalog(serverAddr string) string {
	resourcePath := testResourcePath()
	vmISO := filepath.Join(resourcePath, "cfdev-efi.iso")
	cfISO := filepath.Join(resourcePath, "cf-deps.iso")
	boshISO := filepath.Join(resourcePath, "bosh-deps.iso")

	Expect(vmISO).To(BeAnExistingFile())
	Expect(boshISO).To(BeAnExistingFile())
	Expect(cfISO).To(BeAnExistingFile())

	c := &resource.Catalog{}

	for _, res := range [3]string{vmISO, cfISO, boshISO} {
		name := filepath.Base(res)

		c.Items = append(c.Items,
			resource.Item{
				URL:  fmt.Sprintf("%s/%s", serverAddr, name),
				Name: name,
				MD5:  computeMD5(res),
			},
		)
	}

	bytes, err := json.Marshal(c)
	Expect(err).ToNot(HaveOccurred())

	return string(bytes)
}

func computeMD5(file string) string {
	f, err := os.Open(file)
	Expect(err).ToNot(HaveOccurred())

	defer f.Close()

	h := md5.New()

	_, err = io.Copy(h, f)
	Expect(err).ToNot(HaveOccurred())
	return fmt.Sprintf("%x", h.Sum(nil))
}
