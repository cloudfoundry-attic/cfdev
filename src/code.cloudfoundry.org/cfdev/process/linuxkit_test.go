package process_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
)

var _ = Describe("LinuxKit process", func() {
	var linuxkit process.LinuxKit
	BeforeEach(func() {
		linuxkit = process.LinuxKit{
			Config: config.Config{
				CFDevHome:      "/home-dir/.cfdev",
				StateDir:       "/home-dir/.cfdev/state",
				VpnkitStateDir: "/home-dir/.cfdev/state_vpnkit",
				CacheDir:       "/home-dir/.cfdev/cache",
			},
		}
	})

	It("builds a command", func() {
		start, err := linuxkit.DaemonSpec(4, 4096)
		Expect(err).ToNot(HaveOccurred())

		linuxkitExecPath := "/home-dir/.cfdev/cache/linuxkit"
		Expect(start.Program).To(Equal(linuxkitExecPath))
		Expect(start.ProgramArguments).To(Equal([]string{
			linuxkitExecPath,
			"run", "hyperkit",
			"-console-file",
			"-cpus", "4",
			"-mem", "4096",
			"-hyperkit", "/home-dir/.cfdev/cache/hyperkit",
			"-networking",
			"vpnkit,/home-dir/.cfdev/state_vpnkit/vpnkit_eth.sock,/home-dir/.cfdev/state_vpnkit/vpnkit_port.sock",
			"-fw", "/home-dir/.cfdev/cache/UEFI.fd",
			"-disk", "type=qcow,size=80G,trim=true,qcow-tool=/home-dir/.cfdev/cache/qcow-tool,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144",
			"-disk", "file=/home-dir/.cfdev/cache/cf-deps.iso",
			"-state", "/home-dir/.cfdev/state",
			"--uefi", "/home-dir/.cfdev/cache/cfdev-efi.iso",
		}))
		Expect(start.RunAtLoad).To(BeFalse())
	})

	Context("DepsIsoPath is set", func() {
		Context("DepsIsoPath exists", func() {
			var assetDir string
			var err error
			assetDir, err = ioutil.TempDir(os.TempDir(), "asset")
			var assetUrl = "https://s3.amazonaws.com/cfdev-test-assets/test-deps.dev"

			BeforeEach(func() {
				linuxkit.DepsIsoPath = path.Join(assetDir, "test-deps.dev")
				Expect(err).ToNot(HaveOccurred())
				downloadTestAsset(assetDir, assetUrl)
			})

			It("sets linuxkit to use provided iso", func() {
				start, err := linuxkit.DaemonSpec(4, 4096)
				Expect(err).ToNot(HaveOccurred())

				linuxkitExecPath := "/home-dir/.cfdev/cache/linuxkit"
				Expect(start.Program).To(Equal(linuxkitExecPath))
				Expect(start.ProgramArguments).To(ConsistOf(
					linuxkitExecPath,
					"run", "hyperkit",
					"-console-file",
					"-cpus", "4",
					"-mem", "4096",
					"-hyperkit", "/home-dir/.cfdev/cache/hyperkit",
					"-networking",
					"vpnkit,/home-dir/.cfdev/state_vpnkit/vpnkit_eth.sock,/home-dir/.cfdev/state_vpnkit/vpnkit_port.sock",
					"-fw", "/home-dir/.cfdev/cache/UEFI.fd",
					"-disk", "type=qcow,size=80G,trim=true,qcow-tool=/home-dir/.cfdev/cache/qcow-tool,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144",
					"-disk", "file="+path.Join(assetDir, "test-deps.dev"),
					"-state", "/home-dir/.cfdev/state",
					"--uefi", "/home-dir/.cfdev/cache/cfdev-efi.iso",
				))
			})
		})
		Context("DepsIsoPath does not exist", func() {
			BeforeEach(func() {
				linuxkit.DepsIsoPath = "/some/path/that/does/not/exist"
			})

			It("returns file not found error", func() {
				_, err := linuxkit.DaemonSpec(4, 4096)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func downloadTestAsset(targetDir string, resourceUrl string) error {
	out, err := os.Create(filepath.Join(targetDir, "test-deps.dev"))
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(resourceUrl)
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
