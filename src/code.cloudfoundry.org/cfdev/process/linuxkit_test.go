// +build darwin

package process_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/golang/mock/gomock"
)

var _ = Describe("LinuxKit process", func() {
	var (
		linuxkit       process.LinuxKit
		mockController *gomock.Controller
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())

		linuxkit = process.LinuxKit{
			Config: config.Config{
				CFDevHome:      "/home-dir/.cfdev",
				StateDir:       "/home-dir/.cfdev/state",
				CacheDir:       "/home-dir/.cfdev/cache",
				VpnKitStateDir: "/home-dir/.cfdev/state_vpnkit",
			},
		}
	})

	AfterEach(func(){
		mockController.Finish()
	})

	Context("DepsIsoPath exists", func() {
		var depsIsoPath string
		var tmpDir string

		BeforeEach(func() {
			tmpDir, err := ioutil.TempDir("", "process-test")
			Expect(err).ToNot(HaveOccurred())
			depsIsoPath = filepath.Join(tmpDir, "some-deps-iso")
			_, err = os.Create(depsIsoPath)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("sets linuxkit to use provided iso", func() {
			start, err := linuxkit.DaemonSpec(4, 4096, depsIsoPath)
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
				"-disk", "file="+depsIsoPath,
				"-state", "/home-dir/.cfdev/state",
				"--uefi", "/home-dir/.cfdev/cache/cfdev-efi.iso",
			))
		})

		Context("DepsIsoPath does not exist", func() {
			It("returns file not found error", func() {
				_, err := linuxkit.DaemonSpec(4, 4096, "/some/path/that/does/not/exist")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
