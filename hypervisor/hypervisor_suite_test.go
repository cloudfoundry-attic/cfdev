package hypervisor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"testing"
)

func TestHypervisor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hypervisor Suite")
}

var assetDir string

var _ = BeforeSuite(func() {
	testIsoUrl := "https://s3.amazonaws.com/cfdev-test-assets/test.iso"
	testVHDUrl := "https://s3.amazonaws.com/cfdev-test-assets/test-hd.vhdx"

	var err error
	assetDir, err = ioutil.TempDir("", "hypervtest-assets")
	Expect(err).NotTo(HaveOccurred())

	downloadFile(filepath.Join(assetDir, "cfdev-efi-v2.iso"), testIsoUrl)
	downloadFile(filepath.Join(assetDir, "disk.vhdx"), testVHDUrl)
})

var _ = AfterSuite(func() {
	os.RemoveAll(assetDir)
})

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