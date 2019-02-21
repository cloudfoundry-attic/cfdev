package hypervisor

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"io"
	"time"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type LinuxKit struct {
	Config       config.Config
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
	daemonSpec, err := l.DaemonSpec(vm.CPUs, vm.MemoryMB)
	if err != nil {
		return err
	}
	return l.DaemonRunner.AddDaemon(daemonSpec)
}

func (l *LinuxKit) Start(vmName string) error {
	return l.DaemonRunner.Start(LinuxKitLabel)
}

func (l *LinuxKit) Destroy(vmName string) error {
	return l.DaemonRunner.RemoveDaemon(LinuxKitLabel)
}

func (l *LinuxKit) IsRunning(vmName string) (bool, error) {
	return l.DaemonRunner.IsRunning(LinuxKitLabel)
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
