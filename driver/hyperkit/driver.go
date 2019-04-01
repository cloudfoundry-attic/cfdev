package hyperkit

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	e "code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/pkg/cfdevd/client"
	"code.cloudfoundry.org/cfdev/runner"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type Hyperkit struct {
	UI           driver.UI
	Config       config.Config
	DaemonRunner driver.DaemonRunner
	CFDevD       *client.Client
	SudoShell    *runner.SudoShell
}

func New(
	cfg config.Config,
	daemonRunner driver.DaemonRunner,
	ui driver.UI,
	cfdevdClient *client.Client,
) driver.Driver {
	return &Hyperkit{
		UI:           ui,
		Config:       cfg,
		DaemonRunner: daemonRunner,
		CFDevD:       cfdevdClient,
		SudoShell:    &runner.SudoShell{},
	}
}

func (d *Hyperkit) CheckRequirements() error {
	return nil
}

// This pre-start hook was added to accommodate, albeit hastily,
// a request have any actions that require sudo privileges sooner
// since 'cf dev start' can take some time. For now, only the darwin
// architecture make use of it to since it is not expected to be ran with
// administrator privileges.
// https://github.com/cloudfoundry-incubator/cfdev/issues/41
func (d *Hyperkit) Prestart() error {
	d.UI.Say("Installing cfdevd network helper (requires administrator privileges)...")
	err := d.installCFDevDaemon()
	if err != nil {
		return e.SafeWrap(err, "installing cfdevd")
	}

	d.UI.Say("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")
	_, err = d.CFDevD.AddIPAlias()
	if err != nil {
		return e.SafeWrap(err, "adding network aliases")
	}

	return nil
}

func (d *Hyperkit) Start(cpus int, memory int, efiPath string) error {
	d.UI.Say("Creating the VM...")
	err := d.DaemonRunner.AddDaemon(d.daemonSpec(cpus, memory, efiPath))
	if err != nil {
		return e.SafeWrap(err, "creating the vm")
	}

	d.UI.Say("Starting VPNKit...")
	err = driver.WriteHttpConfig(d.Config)
	if err != nil {
		return e.SafeWrap(err, "setting up networking")
	}

	err = d.DaemonRunner.AddDaemon(d.networkingDaemonSpec())
	if err != nil {
		return e.SafeWrap(err, "starting vpnkit")
	}

	err = d.DaemonRunner.Start(driver.VpnKitLabel)
	if err != nil {
		return e.SafeWrap(err, "starting vpnkit")
	}

	err = d.waitForNetworking()
	if err != nil {
		return e.SafeWrap(err, "connect to vpnkit")
	}

	err = d.DaemonRunner.Start(driver.LinuxKitLabel)
	if err != nil {
		return e.SafeWrap(err, "starting linuxkit")
	}

	return nil
}

func (d *Hyperkit) Stop() error {
	var reterr error

	if err := d.DaemonRunner.Stop(driver.LinuxKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to stop the VM")
	}

	if err := SafeKill(filepath.Join(d.Config.StateLinuxkit, "hyperkit.pid"), "hyperkit"); err != nil {
		reterr = e.SafeWrap(err, "failed to stop the VM")
	}

	if err := d.DaemonRunner.RemoveDaemon(driver.LinuxKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to destroy the VM")
	}

	if err := d.DaemonRunner.Stop(driver.VpnKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := d.DaemonRunner.RemoveDaemon(driver.VpnKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to destroy vpnkit")
	}

	if _, err := d.CFDevD.RemoveIPAlias(); err != nil {
		reterr = e.SafeWrap(err, "failed to remove IP aliases")
	}

	if _, err := d.CFDevD.Uninstall(); err != nil {
		reterr = e.SafeWrap(err, "failed to uninstall cfdevd")
	}

	return reterr
}

func (d *Hyperkit) IsRunning() (bool, error) {
	return d.DaemonRunner.IsRunning(driver.LinuxKitLabel)
}

func (d *Hyperkit) daemonSpec(cpus, mem int, efiPath string) daemon.DaemonSpec {
	var (
		linuxkit       = filepath.Join(d.Config.BinaryDir, "linuxkit")
		hyperkit       = filepath.Join(d.Config.BinaryDir, "hyperkit")
		uefi           = filepath.Join(d.Config.BinaryDir, "UEFI.fd")
		qcowtool       = filepath.Join(d.Config.BinaryDir, "qcow-tool")
		vpnkitEthSock  = filepath.Join(d.Config.VpnKitStateDir, "vpnkit_eth.sock")
		vpnkitPortSock = filepath.Join(d.Config.VpnKitStateDir, "vpnkit_port.sock")
		diskArgs       = []string{
			"type=qcow",
			"size=120G",
			"trim=true",
			fmt.Sprintf("qcow-tool=%s", qcowtool),
			"qcow-onflush=os",
			"qcow-compactafter=262144",
			"qcow-keeperased=262144",
		}
	)

	return daemon.DaemonSpec{
		Label:       driver.LinuxKitLabel,
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
			"-state", d.Config.StateLinuxkit,
			"-uefi",
			efiPath,
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(d.Config.LogDir, "linuxkit.stdout.log"),
		StderrPath: path.Join(d.Config.LogDir, "linuxkit.stderr.log"),
	}
}
