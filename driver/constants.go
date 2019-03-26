package driver

import "code.cloudfoundry.org/cfdev/daemon"

const (
	VMName          = "cfdev"
	VpnKitLabel     = "org.cloudfoundry.cfdev.vpnkit"
	LinuxKitLabel   = "org.cloudfoundry.cfdev.linuxkit"
	ContainerSubnet = "10.144.0.0/16"
)

type UI interface {
	Say(message string, args ...interface{})
}

type DaemonRunner interface {
	AddDaemon(daemon.DaemonSpec) error
	RemoveDaemon(string) error
	Start(string) error
	Stop(string) error
	IsRunning(string) (bool, error)
}

type Driver interface {
	Prestart() error
	Start(cpus int, memory int, efiPath string) error
	Stop() error
	IsRunning() (bool, error)
}