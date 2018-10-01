package network_test

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/network"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VpnKit", func() {
	var (
		tmpDir         string
		vpnkitStateDir string
		vkit           network.VpnKit
		lctl           *daemon.Launchd
		label 		   = "some-vpnkit-label"
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("/var/tmp", "vpnkit-test")
		Expect(err).NotTo(HaveOccurred())
		cacheDir := filepath.Join(tmpDir, "some-cache-dir")
		vpnkitStateDir = filepath.Join(tmpDir, "some-vpnkit-state-dir")
		stateDir := filepath.Join(tmpDir, "some-state-dir")
		homeDir := filepath.Join(tmpDir, "some-home-dir")
		Expect(os.Mkdir(cacheDir, 0777)).To(Succeed())
		Expect(os.Mkdir(vpnkitStateDir, 0777)).To(Succeed())
		Expect(os.Mkdir(stateDir, 0777)).To(Succeed())
		Expect(os.Mkdir(homeDir, 0777)).To(Succeed())
		downloadVpnKit(cacheDir, "https://s3.amazonaws.com/cfdev-ci/vpnkit/vpnkit-darwin-amd64-0.0.0-build.3")
		lctl = &daemon.Launchd{
			PListDir: tmpDir,
		}

		vkit = network.VpnKit{
			Config: config.Config{
				CacheDir:       cacheDir,
				VpnKitStateDir: vpnkitStateDir,
				StateDir:       stateDir,
				CFDevHome:      homeDir,
			},
			DaemonRunner: lctl,
			Label: label,
		}
	})

	AfterEach(func() {
		Expect(lctl.RemoveDaemon(label)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("starts vpnkit", func() {
		Expect(vkit.Start()).To(Succeed())
		conn, err := net.Dial("unix", filepath.Join(vpnkitStateDir, "vpnkit_eth.sock"))
		defer conn.Close()
		Expect(err).NotTo(HaveOccurred())
	})
})

func downloadVpnKit(targetDir string, resourceUrl string) error {
	dest := filepath.Join(targetDir, "vpnkit")
	out, err := os.Create(dest)
	Expect(err).NotTo(HaveOccurred())
	defer out.Close()

	resp, err := http.Get(resourceUrl)
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(os.Chmod(dest, 0777)).To(Succeed())
	return nil
}
