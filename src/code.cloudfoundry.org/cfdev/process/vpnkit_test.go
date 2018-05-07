package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
)

var _ = Describe("VPNKit", func() {
	It("builds a command", func() {
		vpnKit := process.VpnKit{
			Config: config.Config{
				CFDevHome:      "some-home-dir",
				CacheDir:       "some-cache-dir",
				StateDir:       "some-state-dir",
				VpnkitStateDir: "some-vpnkit-state-dir",
			},
		}

		cmd := vpnKit.DaemonSpec()

		Expect(cmd.Program).To(Equal("some-cache-dir/vpnkit"))
		Expect(cmd.ProgramArguments).To(ConsistOf(
			"some-cache-dir/vpnkit",
			"--ethernet",
			"some-vpnkit-state-dir/vpnkit_eth.sock",
			"--port",
			"some-vpnkit-state-dir/vpnkit_port.sock",
			"--vsock-path",
			"some-state-dir/connect",
			"--http",
			"some-vpnkit-state-dir/http_proxy.json",
		))
	})
})
