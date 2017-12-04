package process

import (
	"os/exec"
	"syscall"
)

type LinuxKit struct {
	ImagePath   string
	StatePath   string
	BoshISOPath string
	CFISOPath   string
}

func (s *LinuxKit) Command() *exec.Cmd {
	cmd := exec.Command("linuxkit", "run", "hyperkit",
		"-console-file",
		"-cpus", "4",
		"-mem", "8192",
		"-networking=vpnkit",
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
