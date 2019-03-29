package kvm

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/runner"
	"fmt"
	"path"
	"path/filepath"
	"time"
)

type KVM struct {
	UI           driver.UI
	Config       config.Config
	DaemonRunner driver.DaemonRunner
	SudoShell    *runner.SudoShell
}

func New(
	cfg config.Config,
	daemonRunner driver.DaemonRunner,
	ui driver.UI,
) driver.Driver {
	return &KVM{
		UI:           ui,
		Config:       cfg,
		DaemonRunner: daemonRunner,
		SudoShell:    &runner.SudoShell{},
	}
}

func (d *KVM) CheckRequirements() error {
	// TODO:
	return nil
}

func (d *KVM) Prestart() error {
	// no-op
	return nil
}

func (d *KVM) Start(cpus int, memory int, efiPath string) error {
	var (
		tapDevice  = "cfdevtap0"
		bridgeName = "virbr0"
	)

	d.setupNetworking(tapDevice, bridgeName)

	d.UI.Say("Creating the VM...")
	err := d.DaemonRunner.AddDaemon(d.daemonSpec(cpus, memory, tapDevice, efiPath))
	if err != nil {
		return err
	}

	d.UI.Say("Starting the VM...")
	err = d.DaemonRunner.Start(driver.LinuxKitLabel)
	if err != nil {
		return err
	}

	d.UI.Say("Fetching VM Address...")
	ip, err := d.fetchIP()
	if err != nil {
		return err
	}

	d.setupRoutes(ip)
	return nil
}

func (d *KVM) Stop() error {
	var (
		tapDevice = "cfdevtap0"
	)

	d.DaemonRunner.Stop(driver.LinuxKitLabel)
	d.DaemonRunner.RemoveDaemon(driver.LinuxKitLabel)
	d.teardownRoutes()
	d.teardownNetworking(tapDevice)
	return nil
}

func (d *KVM) IsRunning() (bool, error) {
	return d.DaemonRunner.IsRunning(driver.LinuxKitLabel)
}

func (d *KVM) fetchIP() (string, error) {
	var (
		ticker  = time.NewTicker(time.Second)
		timeout = time.After(2 * time.Minute)
		err     error
	)

	for {
		select {
		case <-ticker.C:
			var ip string
			ip, err = driver.IP(d.Config)
			if err == nil {
				return ip, nil
			}
		case <-timeout:
			return "", err
		}
	}
}

func (d *KVM) daemonSpec(cpus int, mem int, tapDevice, efiPath string) daemon.DaemonSpec {
	var (
		linuxkit = filepath.Join(d.Config.BinaryDir, "linuxkit")
		ovmf     = filepath.Join(d.Config.BinaryDir, "OVMF.fd")
		diskPath = filepath.Join(d.Config.StateLinuxkit, "disk.qcow2")
	)

	return daemon.DaemonSpec{
		Label:   driver.LinuxKitLabel,
		Program: linuxkit,
		ProgramArguments: []string{
			"run", "qemu",
			"-cpus", fmt.Sprintf("%d", cpus),
			"-mem", fmt.Sprintf("%d", mem),
			"-disk", fmt.Sprintf("size=120G,format=qcow2,file=%s", diskPath),
			"-fw", ovmf,
			"-state", d.Config.StateLinuxkit,
			"-networking", fmt.Sprintf("tap,%s", tapDevice),
			"-iso", "-uefi",
			efiPath,
		},
		LogPath: path.Join(d.Config.LogDir, "linuxkit.log"),
	}
}
