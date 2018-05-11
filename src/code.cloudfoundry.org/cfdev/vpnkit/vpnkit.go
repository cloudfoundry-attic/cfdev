package vpnkit

import (
	"net"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/process"

	launchdModels "code.cloudfoundry.org/cfdevd/launchd/models"
)

type Launchd interface {
	AddDaemon(launchdModels.DaemonSpec) error
	Start(label string) error
}

const retries = 5

func Start(config config.Config, launchd Launchd) error {
	vpnKit := process.VpnKit{
		Config: config,
	}
	if err := vpnKit.SetupVPNKit(); err != nil {
		return errors.SafeWrap(err, "Failed to setup VPNKit")
	}
	if err := launchd.AddDaemon(vpnKit.DaemonSpec()); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}
	if err := launchd.Start(process.VpnKitLabel); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}
	attempt := 0
	for {
		conn, err := net.Dial("unix", filepath.Join(config.VpnkitStateDir, "vpnkit_eth.sock"))
		if err == nil {
			conn.Close()
			return nil
		} else if attempt >= retries {
			return errors.SafeWrap(err, "conenct to vpnkit")
		} else {
			time.Sleep(time.Second)
			attempt++
		}
	}
}
