package process

import (
	"os/exec"
	"syscall"
)

type LinuxKit struct {
	ImagePath   string
	StatePath   string
	BoshISOPath string
}

func (s *LinuxKit) Command() *exec.Cmd {
	cmd := exec.Command("linuxkit", "run", "hyperkit",
		"-console-file",
		"-networking=vpnkit",
		"-disk", "size=10G",
		"-disk", "file="+s.BoshISOPath,
		"-state", s.StatePath,
		"--uefi", s.ImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
