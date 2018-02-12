package process

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type LinuxKit struct {
	ExecutablePath      string
	StatePath           string
	HomeDir             string
	OSImagePath         string
	DependencyImagePath string
}

func (s *LinuxKit) Command() *exec.Cmd {
	linuxkit := filepath.Join(s.ExecutablePath, "linuxkit")
	hyperkit := filepath.Join(s.ExecutablePath, "hyperkit")
	uefi := filepath.Join(s.ExecutablePath, "UEFI.fd")
	qcowtool := filepath.Join(s.ExecutablePath, "qcow-tool")
	vpnkitEthSock := filepath.Join(s.HomeDir, "vpnkit_eth.sock")
	vpnkitPortSock := filepath.Join(s.HomeDir, "vpnkit_port.sock")

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
		"-networking", fmt.Sprintf("vpnkit,%v,%v", vpnkitEthSock, vpnkitPortSock),
		"-fw", uefi,
		"-disk", strings.Join(diskArgs, ","),
		"-disk", "file="+s.DependencyImagePath,
		"-state", s.StatePath,
		"--uefi",
		s.OSImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
