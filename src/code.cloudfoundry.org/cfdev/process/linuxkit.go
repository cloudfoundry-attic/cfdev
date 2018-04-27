package process

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"code.cloudfoundry.org/cfdev/config"
	"io"
	"os"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type LinuxKit struct {
	Config config.Config
	DepsIsoPath string
}

func (l *LinuxKit) Command(cpus, mem int) (*exec.Cmd, error) {
	linuxkit := filepath.Join(l.Config.CacheDir, "linuxkit")
	hyperkit := filepath.Join(l.Config.CacheDir, "hyperkit")
	uefi := filepath.Join(l.Config.CacheDir, "UEFI.fd")
	qcowtool := filepath.Join(l.Config.CacheDir, "qcow-tool")
	vpnkitEthSock := filepath.Join(l.Config.CFDevHome, "vpnkit_eth.sock")
	vpnkitPortSock := filepath.Join(l.Config.CFDevHome, "vpnkit_port.sock")

	if l.DepsIsoPath == "" {
		l.DepsIsoPath = filepath.Join(l.Config.CacheDir, "cf-deps.iso")
	}

	if _, err := os.Stat(l.DepsIsoPath); os.IsNotExist(err) {
		return nil, err
	}

	dependencyImagePath := l.DepsIsoPath
	osImagePath := filepath.Join(l.Config.CacheDir, "cfdev-efi.iso")

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
		"-cpus", fmt.Sprintf("%d", cpus),
		"-mem", fmt.Sprintf("%d", mem),
		"-hyperkit", hyperkit,
		"-networking", fmt.Sprintf("vpnkit,%v,%v", vpnkitEthSock, vpnkitPortSock),
		"-fw", uefi,
		"-disk", strings.Join(diskArgs, ","),
		"-disk", "file="+dependencyImagePath,
		"-state", l.Config.StateDir,
		"--uefi",
		osImagePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd, nil
}
