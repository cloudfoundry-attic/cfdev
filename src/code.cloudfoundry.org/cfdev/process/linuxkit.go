package process

import (
	"os/exec"
	"syscall"
)

type LinuxKit struct {
	ImagePath string
	StatePath string
}

func (s *LinuxKit) Command() *exec.Cmd {
	cmd := exec.Command("linuxkit", "run", "hyperkit",
		"-networking=vpnkit",
		"-disk", "size=10G",
		"-state", s.StatePath,
		"--uefi", s.ImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
