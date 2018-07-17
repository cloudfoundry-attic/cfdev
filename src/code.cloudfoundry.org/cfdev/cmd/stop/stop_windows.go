package stop

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdevd/launchd"
	"github.com/spf13/cobra"
	"os/exec"
)


func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	var reterr error

	if err := s.HyperV.Stop("cfdev"); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop the VM")
	}

	if err := s.Launchd.Stop(daemonSpec(process.VpnKitLabel, s.Config.CFDevHome)); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := s.Launchd.RemoveDaemon(daemonSpec(process.VpnKitLabel, s.Config.CFDevHome)); err != nil {
		reterr = errors.SafeWrap(err, "failed to uninstall vpnkit")
	}

	if err := s.HostNet.RemoveNetworkSwitch(); err != nil {
		reterr = errors.SafeWrap(err, "failed to remove network switch")
	}

	registryDeleteCmd := `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
		`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
		`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`

	if err := exec.Command("powershell.exe", "-Command", registryDeleteCmd).Run(); err != nil {
		reterr = errors.SafeWrap(err, "failed to remove service registries")
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}
	return nil
}

func daemonSpec(label, cfdevHome string) launchd.DaemonSpec {
	return launchd.DaemonSpec{
		Label:     label,
		CfDevHome: cfdevHome,
	}
}
