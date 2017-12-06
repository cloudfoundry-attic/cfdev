package process

import (
	"os/exec"
	"path/filepath"
	"syscall"
)

type LinuxKit struct {
	ExecutablePath string
	ImagePath      string
	StatePath      string
	BoshISOPath    string
	CFISOPath      string
}

func (s *LinuxKit) Command() *exec.Cmd {
	linuxkit := filepath.Join(s.ExecutablePath, "linuxkit")
	hyperkit := filepath.Join(s.ExecutablePath, "hyperkit")
	uefi := filepath.Join(s.ExecutablePath, "UEFI.fd")
	vpnkit := filepath.Join(s.ExecutablePath, "vpnkit")

	cmd := exec.Command(linuxkit, "run", "hyperkit",
		"-console-file",
		"-cpus", "4",
		"-mem", "8192",
		"-hyperkit", hyperkit,
		"-networking", "vpnkit",
		"-vpnkit", vpnkit,
		"-fw", uefi,
		"-disk", "size=50G",
		"-disk", "file="+s.BoshISOPath,
		"-disk", "file="+s.CFISOPath,
		"-state", s.StatePath,
		"--uefi", s.ImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
