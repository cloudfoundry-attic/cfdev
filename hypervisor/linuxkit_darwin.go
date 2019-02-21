package hypervisor

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func (l *LinuxKit) Stop(vmName string) error {
	var reterr error
	if err := l.DaemonRunner.Stop(LinuxKitLabel); err != nil {
		reterr = err
	}
	if err := SafeKill(
		filepath.Join(l.Config.StateLinuxkit, "hyperkit.pid"),
		"hyperkit",
	); err != nil {
		reterr = err
	}
	return reterr
}

func (l *LinuxKit) DaemonSpec(cpus, mem int) (daemon.DaemonSpec, error) {
	var (
		linuxkit       = filepath.Join(l.Config.BinaryDir, "linuxkit")
		hyperkit       = filepath.Join(l.Config.BinaryDir, "hyperkit")
		uefi           = filepath.Join(l.Config.BinaryDir, "UEFI.fd")
		qcowtool       = filepath.Join(l.Config.BinaryDir, "qcow-tool")
		osImagePath    = filepath.Join(l.Config.BinaryDir, "cfdev-efi-v2.iso")
		vpnkitEthSock  = filepath.Join(l.Config.VpnKitStateDir, "vpnkit_eth.sock")
		vpnkitPortSock = filepath.Join(l.Config.VpnKitStateDir, "vpnkit_port.sock")
		diskArgs       = []string{
			"type=qcow",
			"size=120G",
			"trim=true",
			fmt.Sprintf("qcow-tool=%s", qcowtool),
			"qcow-onflush=os",
			"qcow-compactafter=262144",
			"qcow-keeperased=262144",
		}
	)

	return daemon.DaemonSpec{
		Label:       LinuxKitLabel,
		Program:     linuxkit,
		SessionType: "Background",
		ProgramArguments: []string{
			linuxkit, "run", "hyperkit",
			"-console-file",
			"-cpus", fmt.Sprintf("%d", cpus),
			"-mem", fmt.Sprintf("%d", mem),
			"-hyperkit", hyperkit,
			"-networking", fmt.Sprintf("vpnkit,%v,%v", vpnkitEthSock, vpnkitPortSock),
			"-fw", uefi,
			"-disk", strings.Join(diskArgs, ","),
			"-state", l.Config.StateLinuxkit,
			"-uefi",
			osImagePath,
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(l.Config.LogDir, "linuxkit.stdout.log"),
		StderrPath: path.Join(l.Config.LogDir, "linuxkit.stderr.log"),
	}, nil
}
