package network

import (
	"net"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cfdev/errors"

	"path"

	"code.cloudfoundry.org/cfdev/daemon"
)

const retries = 5

func (v *VpnKit) Start() error {
	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to Setup VPNKit")
	}
	if err := v.DaemonRunner.AddDaemon(v.daemonSpec()); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}
	if err := v.DaemonRunner.Start(v.Label); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}
	attempt := 0
	for {
		conn, err := net.Dial("unix", filepath.Join(v.Config.VpnKitStateDir, "vpnkit_eth.sock"))
		if err == nil {
			conn.Close()
			return nil
		} else if attempt >= retries {
			return errors.SafeWrap(err, "connect to vpnkit")
		} else {
			time.Sleep(time.Second)
			attempt++
		}
	}
}

func (v *VpnKit) Destroy() error {
	return v.DaemonRunner.RemoveDaemon(v.Label)
}

func (v *VpnKit) daemonSpec() daemon.DaemonSpec {
	return daemon.DaemonSpec{
		Label:       v.Label,
		Program:     path.Join(v.Config.CacheDir, "vpnkit"),
		SessionType: "Background",
		ProgramArguments: []string{
			path.Join(v.Config.CacheDir, "vpnkit"),
			"--ethernet", path.Join(v.Config.VpnKitStateDir, "vpnkit_eth.sock"),
			"--port", path.Join(v.Config.VpnKitStateDir, "vpnkit_port.sock"),
			"--vsock-path", path.Join(v.Config.StateDir, "connect"),
			"--http", path.Join(v.Config.VpnKitStateDir, "http_proxy.json"),
			"--host-names", "host.cfdev.sh",
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(v.Config.CFDevHome, "vpnkit.stdout.log"),
		StderrPath: path.Join(v.Config.CFDevHome, "vpnkit.stderr.log"),
	}
}

func (v *VpnKit) Setup() error {
	return v.writeHttpConfig()
}
