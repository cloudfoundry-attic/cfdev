package process

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"code.cloudfoundry.org/cfdev/config"
)

type LinuxKit struct {
	Config config.Config
}

func (s *LinuxKit) Command() *exec.Cmd {
	linuxkit := filepath.Join(s.Config.CacheDir, "linuxkit")
	hyperkit := filepath.Join(s.Config.CacheDir, "hyperkit")
	uefi := filepath.Join(s.Config.CacheDir, "UEFI.fd")
	qcowtool := filepath.Join(s.Config.CacheDir, "qcow-tool")
	vpnkitEthSock := filepath.Join(s.Config.CFDevHome, "vpnkit_eth.sock")
	vpnkitPortSock := filepath.Join(s.Config.CFDevHome, "vpnkit_port.sock")
	dependencyImagePath := filepath.Join(s.Config.CacheDir, "cf-oss-deps.iso")
	osImagePath := filepath.Join(s.Config.CacheDir, "cfdev-efi.iso")

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
		"-disk", "file="+dependencyImagePath,
		"-state", s.Config.StateDir,
		"--uefi",
		osImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}
