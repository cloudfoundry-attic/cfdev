package hyperv

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (d *HyperV) createVM(name string, cpus int, memory int, efiPath string) error {
	var (
		cfDevVHD    = filepath.Join(d.Config.StateDir, "disk.vhdx")
	)

	command := fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", name)
	_, err := d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("creating new vm: %s", err)
	}

	command = fmt.Sprintf("Set-VM -Name %s "+
		"-AutomaticStartAction Nothing "+
		"-AutomaticStopAction ShutDown "+
		"-CheckpointType Disabled "+
		fmt.Sprintf("-MemoryStartupBytes %dMB ", memory)+
		"-StaticMemory "+
		fmt.Sprintf("-ProcessorCount %d", cpus),
		name)
	_, err = d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting vm properites (memoryMB:%d, cpus:%d): %s", memory, cpus, err)
	}

	command = fmt.Sprintf(`Add-VMDvdDrive -VMName %s -Path "%s"`, name, efiPath)
	_, err = d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", efiPath, err)
	}

	command = fmt.Sprintf("Remove-VMNetworkAdapter -VMName %s", name)
	_, err = d.Powershell.Output(command)
	if err != nil {
		fmt.Printf("failed to remove network adapter: %s", err)
	}

	command = fmt.Sprintf("Add-VMHardDiskDrive -VMName %s "+
		`-Path "%s"`, name, cfDevVHD)
	_, err = d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("adding vhd %s : %s", cfDevVHD, err)
	}

	command = fmt.Sprintf("Set-VMFirmware "+
		"-VMName %s "+
		"-EnableSecureBoot Off "+
		"-FirstBootDevice $cdrom",
		name)
	_, err = d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting firmware: %s", err)
	}

	command = fmt.Sprintf("Set-VMComPort "+
		"-VMName %s "+
		"-number 1 "+
		"-Path \\\\.\\pipe\\cfdev-com",
		name)
	_, err = d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting com port: %s", err)
	}

	return nil
}

func (d *HyperV) start(vmName string) error {
	if exists, err := d.exists(vmName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("hyperv vm with name %s does not exist", vmName)
	}

	command := fmt.Sprintf("Start-VM -Name %s", vmName)
	if _, err := d.Powershell.Output(command); err != nil {
		return fmt.Errorf("start-vm: %s", err)
	}

	return nil
}

func (d *HyperV) stop(vmName string) error {
	if exists, err := d.exists(vmName); err != nil {
		return err
	} else if !exists {
		return nil
	}

	command := fmt.Sprintf("Stop-VM -Name %s -Turnoff", vmName)
	if _, err := d.Powershell.Output(command); err != nil {
		return fmt.Errorf("stopping vm: %s", err)
	}

	return nil
}

func (d *HyperV) destroy(vmName string) error {
	if exists, err := d.exists(vmName); err != nil {
		return err
	} else if !exists {
		return nil
	}

	command := fmt.Sprintf("Remove-VM -Name %s -Force", vmName)
	if _, err := d.Powershell.Output(command); err != nil {
		return fmt.Errorf("removing vm: %s", err)
	}

	return nil
}

func (d *HyperV) isRunning(vmName string) (bool, error) {
	if exists, err := d.exists(vmName); err != nil || !exists {
		return false, err
	}

	command := fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName)
	output, err := d.Powershell.Output(command)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(output), "Running") {
		return true, nil
	}

	return false, nil
}

func (d *HyperV) exists(vmName string) (bool, error) {
	command := fmt.Sprintf("Get-VM -Name %s*", vmName)
	output, err := d.Powershell.Output(command)
	if err != nil {
		return false, fmt.Errorf("getting vms: %s", err)
	}

	return output != "", nil
}
