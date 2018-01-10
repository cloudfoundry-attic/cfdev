package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/process"
)

var _ = Describe("LinuxKit process", func() {
	It("builds a command", func() {
		linuxkit := process.LinuxKit{
			ExecutablePath:      "/home-dir/.cfdev/cache",
			OSImagePath:         "/home-dir/.cfdev/image",
			StatePath:           "/home-dir/.cfdev/state",
			DependencyImagePath: "/home-dir/.cfdev/bosh-cf-deps.iso",
		}

		start := linuxkit.Command()

		linuxkitExecPath := "/home-dir/.cfdev/cache/linuxkit"
		Expect(start.Path).To(Equal(linuxkitExecPath))
		Expect(start.Args).To(ConsistOf(
			linuxkitExecPath,
			"run", "hyperkit",
			"-console-file",
			"-cpus", "4",
			"-mem", "8192",
			"-hyperkit", "/home-dir/.cfdev/cache/hyperkit",
			"-networking", "vpnkit",
			"-fw", "/home-dir/.cfdev/cache/UEFI.fd",
			"-vpnkit", "/home-dir/.cfdev/cache/vpnkit",
			"-disk", "type=qcow,size=50G,trim=true,qcow-tool=/home-dir/.cfdev/cache/qcow-tool,qcow-onflush=os,qcow-compactafter=262144,qcow-keeperased=262144",
			"-disk", "file=/home-dir/.cfdev/bosh-cf-deps.iso",
			"-state", "/home-dir/.cfdev/state",
			"--uefi", "/home-dir/.cfdev/image",
		))
		Expect(start.SysProcAttr.Setpgid).To(BeTrue())
	})
})
