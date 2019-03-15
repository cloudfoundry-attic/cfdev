package hypervisor_test

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/hypervisor"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LinuxKit process", func() {
	var (
		linuxkit       hypervisor.LinuxKit
		mockController *gomock.Controller
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())

		linuxkit = hypervisor.LinuxKit{
			Config: config.Config{
				CFDevHome:      "/home-dir/.cfdev",
				StateDir:       "/home-dir/.cfdev/state",
				CacheDir:       "/home-dir/.cfdev/cache",
				BinaryDir:      "/home-dir/.cfdev/bin",
				StateLinuxkit:  "/home-dir/.cfdev/state/linuxkit",
				VpnKitStateDir: "/home-dir/.cfdev/state_vpnkit",
			},
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	It("sets linuxkit to use provided iso", func() {
		start, err := linuxkit.DaemonSpec(4, 4096)
		Expect(err).ToNot(HaveOccurred())

		linuxkitExecPath := "/home-dir/.cfdev/bin/linuxkit"
		Expect(start.Program).To(Equal(linuxkitExecPath))
		Expect(start.ProgramArguments).To(ConsistOf(
			"run", "qemu",
			"-cpus", "4",
			"-mem", "4096",
			"-disk", "size=120G,format=qcow2,file=/home-dir/.cfdev/state/linuxkit/disk.qcow2",
			"-fw", "/home-dir/.cfdev/bin/OVMF.fd",
			"-state", "/home-dir/.cfdev/state/linuxkit",
			"-networking", "tap,cfdevtap0",
			"-iso", "-uefi",
			"/home-dir/.cfdev/bin/cfdev-efi-v2.iso",
		))
	})
})
