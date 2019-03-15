package hyperv

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/driver"
	e "code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/runner"
)

type HyperV struct {
	UI            driver.UI
	Config        config.Config
	DaemonRunner  driver.DaemonRunner
	Powershell    runner.Powershell
	EthernetGUID  string
	PortGUID      string
	ForwarderGUID string
}

func New(
	cfg config.Config,
	daemonRunner driver.DaemonRunner,
	ui driver.UI,
	powershell runner.Powershell,
	ethernetGUID string,
	portGUID string,
	forwarderGUID string,
) *HyperV {
	return &HyperV{
		UI:            ui,
		Config:        cfg,
		DaemonRunner:  daemonRunner,
		Powershell:    powershell,
		EthernetGUID:  ethernetGUID,
		PortGUID:      portGUID,
		ForwarderGUID: forwarderGUID,
	}
}

func (d *HyperV) Prestart() error {
	// no-op
	return nil
}

func (d *HyperV) Start(cpus int, memory int, efiPath string) error {
	d.UI.Say("Creating the VM...")
	err := d.createVM(driver.VMName, cpus, memory, efiPath)
	if err != nil {
		return e.SafeWrap(err, "creating the vm")
	}

	d.UI.Say("Starting VPNKit...")
	vmGUID, err := d.setupNetworking()
	if err != nil {
		return e.SafeWrap(err, "setting up networking")
	}

	err = d.DaemonRunner.AddDaemon(d.networkingDaemonSpec(driver.VpnKitLabel, vmGUID))
	if err != nil {
		return e.SafeWrap(err, "starting vpnkit")
	}

	err = d.DaemonRunner.Start(driver.VpnKitLabel)
	if err != nil {
		return e.SafeWrap(err, "starting vpnkit")
	}

	d.UI.Say("Starting the VM...")
	if err := d.start(driver.VMName); err != nil {
		return e.SafeWrap(err, "starting the vm")
	}

	err = d.AddLoopbackAliases(driver.VMName, d.Config.BoshDirectorIP, d.Config.CFRouterIP)
	if err != nil {
		return e.SafeWrap(err, "adding network aliases")
	}

	return nil
}

func (d *HyperV) Stop() error {
	var reterr error

	if err := d.stop(driver.VMName); err != nil {
		reterr = e.SafeWrap(err, "failed to stop the VM")
	}

	if err := d.destroy(driver.VMName); err != nil {
		reterr = e.SafeWrap(err, "failed to destroy the VM")
	}

	if err := d.DaemonRunner.Stop(driver.VpnKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to stop vpnkit")
	}

	if err := d.DaemonRunner.RemoveDaemon(driver.VpnKitLabel); err != nil {
		reterr = e.SafeWrap(err, "failed to stop vpnkit")
	}

	_, err := d.Powershell.Output(registryDeleteCmd)
	if err != nil {
		return e.SafeWrap(err, "failed to remove network registry entries")
	}

	err = d.RemoveLoopbackAliases(driver.VMName, d.Config.BoshDirectorIP, d.Config.CFRouterIP)
	if err != nil {
		reterr = e.SafeWrap(err, "failed to remove IP aliases")
	}

	return reterr
}

func (d *HyperV) IsRunning() (bool, error) {
	return d.isRunning(driver.VMName)
}
