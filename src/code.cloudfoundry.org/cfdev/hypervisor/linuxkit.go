package hypervisor

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
	DaemonRunner DaemonRunner
}

type DaemonRunner interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(string) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

const LinuxKitLabel = "org.cloudfoundry.cfdev.linuxkit"

func (l *LinuxKit) CreateVM(vm VM) error {
	daemonSpec, err := l.DaemonSpec(vm.CPUs, vm.MemoryMB, vm.DepsIso)
	if err != nil {
		return err
	}
	return l.DaemonRunner.AddDaemon(daemonSpec)
}

func (l *LinuxKit) Start(vmName string) error {
	return l.DaemonRunner.Start(LinuxKitLabel)
}

func (l *LinuxKit) Stop(vmName string) error {
	var reterr error
	if err := l.DaemonRunner.Stop(LinuxKitLabel); err != nil {
		reterr = err
	}
	if err := SafeKill(
		filepath.Join(l.Config.StateDir, "hyperkit.pid"),
		"hyperkit",
	); err != nil {
		reterr = err
	}
	return reterr
}

func (l *LinuxKit) Destroy(vmName string) error {
	return l.DaemonRunner.RemoveDaemon(LinuxKitLabel)
}

func (l *LinuxKit) IsRunning(vmName string) (bool, error) {
	return l.DaemonRunner.IsRunning(LinuxKitLabel)
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
			running, err := l.DaemonRunner.IsRunning(LinuxKitLabel)
			if !running && err == nil {
				exit <- "linuxkit"
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
