package process

const VpnKitLabel = "org.cloudfoundry.cfdev.vpnkit"

func (v *VpnKit) Stop() error {
	return v.DaemonRunner.Stop(VpnKitLabel)
}