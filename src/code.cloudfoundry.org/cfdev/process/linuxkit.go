package process

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
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
	qcowtool := filepath.Join(s.ExecutablePath, "qcow-tool")

	diskArgs := []string{
		"type=qcow",
		"size=50G",
		"trim=true",
		fmt.Sprintf("qcow-tool=%s", qcowtool),
		"qcow-onflush=os",
		"qcow-compactafter=262144",
		"qcow-keeperased=262144",
	}

	cmd := exec.Command(linuxkit, "run", "hyperkit",
		"-console-file",
		"-cpus", "4",
		"-mem", "8192",
		"-hyperkit", hyperkit,
		"-networking", "vpnkit",
		"-vpnkit", vpnkit,
		"-fw", uefi,
		"-disk", strings.Join(diskArgs, ","),
		"-disk", "file="+s.BoshISOPath,
		"-disk", "file="+s.CFISOPath,
		"-state", s.StatePath,
		"--uefi", s.ImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
