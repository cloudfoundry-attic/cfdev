package hypervisor

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"fmt"
	"path"
	"path/filepath"
)

func (l *LinuxKit) Stop(vmName string) error {
	return l.DaemonRunner.Stop(LinuxKitLabel)
}

func (l *LinuxKit) DaemonSpec(cpus, mem int) (daemon.DaemonSpec, error) {
	var (
		linuxkit    = filepath.Join(l.Config.BinaryDir, "linuxkit")
		ovmf        = filepath.Join(l.Config.BinaryDir, "OVMF.fd")
		osImagePath = filepath.Join(l.Config.BinaryDir, "cfdev-efi-v2.iso")
		diskPath    = filepath.Join(l.Config.StateLinuxkit, "disk.qcow2")
	)

	return daemon.DaemonSpec{
		Label:   LinuxKitLabel,
		Program: linuxkit,
		ProgramArguments: []string{
			"run", "qemu",
			"-cpus", fmt.Sprintf("%d", cpus),
			"-mem", fmt.Sprintf("%d", mem),
			"-disk", fmt.Sprintf("size=120G,format=qcow2,file=%s", diskPath),
			"-fw", ovmf,
			"-state", l.Config.StateLinuxkit,
			"-networking", "tap,cfdevtap0",
			"-iso", "-uefi",
			osImagePath,
		},
		LogPath: path.Join(l.Config.LogDir, "linuxkit.log"),
	}, nil
}
