// +build windows

package process_test

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("HyperV process", func() {
	var (
		hyperV     process.HyperV
		cfDevHome  string
		testIsoUrl = "https://s3.amazonaws.com/cfdev-test-assets/test.iso"
		err        error
	)

	BeforeEach(func() {
		cfDevHome, err = ioutil.TempDir("", "hypervtest")
		if err != nil {
			log.Fatal(err)
		}

		hyperV = process.HyperV{
			Config: config.Config{
				CFDevHome: cfDevHome,
				CacheDir:  filepath.Join(cfDevHome, "cache"),
			},
		}

		err = os.MkdirAll(hyperV.Config.CacheDir, 0666)
		Expect(err).ToNot(HaveOccurred())

		downloadAssets(hyperV.Config.CacheDir, testIsoUrl)
	})

	AfterEach(func() {
		cmd := exec.Command("powershell.exe", "-Command", "Remove-VM -Name cfdev -Force")
		err := cmd.Run()
		Expect(err).ToNot(HaveOccurred())

		err = os.RemoveAll(cfDevHome)
		Expect(err).ToNot(HaveOccurred())
	})

	It("creates hyperv VM", func() {
		Expect(hyperV.CreateVM("")).To(Succeed())

		cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 10, 1).Should(gexec.Exit())

		cmd = exec.Command("powershell.exe", "-Command", "Get-VMHardDiskDrive -VMName cfdev")
		output, err := cmd.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).ToNot(BeEmpty())
	})
})

func downloadAssets(cacheDir string, isoSource string) {
	downloadFile(filepath.Join(cacheDir, "cfdev-efi.iso"), isoSource)
	downloadFile(filepath.Join(cacheDir, "cf-deps.iso"), isoSource)
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
