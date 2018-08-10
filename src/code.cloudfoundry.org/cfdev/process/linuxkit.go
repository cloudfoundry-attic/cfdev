package process

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type LinuxKit struct {
	Config  config.Config
	Launchd Launchd
}

type Launchd interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(daemon.DaemonSpec) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

const LinuxKitLabel = "org.cloudfoundry.cfdev.linuxkit"

func (l *LinuxKit) CreateVM(VM) error {
	return nil
}

func (l *LinuxKit) Start(cpus int, mem int, depsIsoPath string) error {
	daemonSpec, err := l.DaemonSpec(cpus, mem, depsIsoPath)
	if err != nil {
		return err
	}
	if err := l.Launchd.AddDaemon(daemonSpec); err != nil {
		return err
	}
	return l.Launchd.Start(daemonSpec)
}

func (l *LinuxKit) Stop() error{
	var reterr error
	if err := l.Launchd.Stop(LinuxKitLabel); err != nil {
		reterr = err
	}
	procManager := &Manager{}
	if err := procManager.SafeKill(
		filepath.Join(l.Config.StateDir, "hyperkit.pid"),
		"hyperkit",
	); err != nil {
		reterr = err
	}
	return reterr
}

func (l *LinuxKit) Destroy() error {
	return l.Launchd.RemoveDaemon(LinuxKitLabel)
}

func (l *LinuxKit) IsRunning() (bool, error) {
	return l.Launchd.IsRunning(LinuxKitLabel)
}

func (l *LinuxKit) DaemonSpec(cpus, mem int, depsIsoPath string) (daemon.DaemonSpec, error) {
	linuxkit := filepath.Join(l.Config.CacheDir, "linuxkit")
	hyperkit := filepath.Join(l.Config.CacheDir, "hyperkit")
	uefi := filepath.Join(l.Config.CacheDir, "UEFI.fd")
	qcowtool := filepath.Join(l.Config.CacheDir, "qcow-tool")
	vpnkitEthSock := filepath.Join(l.Config.VpnKitStateDir, "vpnkit_eth.sock")
	vpnkitPortSock := filepath.Join(l.Config.VpnKitStateDir, "vpnkit_port.sock")

	if _, err := os.Stat(depsIsoPath); os.IsNotExist(err) {
		return daemon.DaemonSpec{}, err
	}

	osImagePath := filepath.Join(l.Config.CacheDir, "cfdev-efi.iso")

	diskArgs := []string{
		"type=qcow",
		"size=80G",
		"trim=true",
		fmt.Sprintf("qcow-tool=%s", qcowtool),
		"qcow-onflush=os",
		"qcow-compactafter=262144",
		"qcow-keeperased=262144",
	}

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
			"-disk", "file=" + depsIsoPath,
			"-state", l.Config.StateDir,
			"--uefi",
			osImagePath,
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(l.Config.CFDevHome, "linuxkit.stdout.log"),
		StderrPath: path.Join(l.Config.CFDevHome, "linuxkit.stderr.log"),
	}, nil
}

func (l *LinuxKit) Watch(exit chan string) {
	go func() {
		for {
			running, err := l.Launchd.IsRunning(LinuxKitLabel)
			if !running && err == nil {
				exit <- "linuxkit"
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
