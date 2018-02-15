package process_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/config"
)

var _ = Describe("VPNKit", func() {
	It("builds a command", func() {
		homeDir := "/home"
		cacheDir := "/home/cache"
		stateDir := "/home/state"

		vpnKit := process.VpnKit{
			Config: config.Config{
				CFDevHome:  homeDir,
				CacheDir: cacheDir,
				StateDir: stateDir,
			},
		}

		cmd := vpnKit.Command()

		Expect(cmd.Args).To(ConsistOf(
			"/home/cache/vpnkit",
			"--ethernet",
			"/home/vpnkit_eth.sock",
			"--port",
			"/home/vpnkit_port.sock",
			"--vsock-path",
			"/home/state/connect",
			"--http",
			"/home/http_proxy.json",
		))
	})
})
