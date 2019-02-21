package network

import (
	"code.cloudfoundry.org/cfdev/daemon"
)

func (v *VpnKit) Start() error {
	return nil
}

func (v *VpnKit) Destroy() error {
	return nil
}

func (v *VpnKit) Stop() error {
	return nil
}

func (v *VpnKit) daemonSpec() daemon.DaemonSpec {
	return daemon.DaemonSpec{}
}

func (v *VpnKit) Setup() error {
	return nil
}
