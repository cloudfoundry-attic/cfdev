package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
)

var _ = Describe("LinuxKit process", func() {
	It("builds a command", func() {
		linuxkit := process.LinuxKit{
			Config: config.Config{
				CFDevHome: "/home-dir/.cfdev",
				StateDir:  "/home-dir/.cfdev/state",
				CacheDir:  "/home-dir/.cfdev/cache",
			},
		}

		start := linuxkit.Command(4, 4096)

		linuxkitExecPath := "/home-dir/.cfdev/cache/linuxkit"
		Expect(start.Path).To(Equal(linuxkitExecPath))
		Expect(start.Args).To(ConsistOf(
			linuxkitExecPath,
			"run", "hyperkit",
			"-console-file",
			"-cpus", "4",
			"-mem", "4096",
			"-hyperkit", "/home-dir/.cfdev/cache/hyperkit",
			"-networking",
			"vpnkit,/home-dir/.cfdev/vpnkit_eth.sock,/home-dir/.cfdev/vpnkit_port.sock",
			"-fw", "/home-dir/.cfdev/cache/UEFI.fd",
			"-disk", "type=qcow,size=50G,trim=true,qcow-tool=/home-dir/.cfdev/cache/qcow-tool,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144",
			"-disk", "file=/home-dir/.cfdev/cache/cf-oss-deps.iso",
			"-state", "/home-dir/.cfdev/state",
			"--uefi", "/home-dir/.cfdev/cache/cfdev-efi.iso",
		))
		Expect(start.SysProcAttr.Setpgid).To(BeTrue())
	})
})
