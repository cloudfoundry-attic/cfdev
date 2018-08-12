package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
)

type HyperV struct {
	Config config.Config
}

type VM struct {
	DepsIso  string
	MemoryMB int
	CPUs     int
}

func (h *HyperV) CreateVM(vm VM) error {
	var vmName = "cfdev"
	var cfdevEfiIso = filepath.Join(h.Config.CacheDir, "cfdev-efi.iso")
	if vm.DepsIso == "" {
		vm.DepsIso = filepath.Join(h.Config.CacheDir, "cf-deps.iso")
	}
	var cfDevVHD = filepath.Join(h.Config.CFDevHome, "cfdev.vhd")

	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("creating new vm: %s", err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Set-VM -Name %s "+
		"-AutomaticStartAction Nothing "+
		"-AutomaticStopAction ShutDown "+
		"-CheckpointType Disabled "+
		fmt.Sprintf("-MemoryStartupBytes %dMB ", vm.MemoryMB)+
		"-StaticMemory "+
		fmt.Sprintf("-ProcessorCount %d", vm.CPUs),
		vmName))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("setting vm properites (memoryMB:%d, cpus:%d): %s", vm.MemoryMB, vm.CPUs, err)
	}

	err = addVhdDrive(cfdevEfiIso, vmName)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", cfdevEfiIso, err)
	}

	err = addVhdDrive(vm.DepsIso, vmName)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", vm.DepsIso, err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VMNetworkAdapter "+
		"-VMName %s "+
		"-Name 'Network Adapter'",
		vmName))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("removing netowrk adapter: %s", err)
	}

	if _, err := os.Stat(cfDevVHD); err == nil {
		err := os.RemoveAll(cfDevVHD)
		if err != nil {
			return fmt.Errorf("removing any vhds: %s", err)
		}
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VHD -Path %s "+
		"-SizeBytes '200000000000' "+
		"-Dynamic", cfDevVHD))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("creating new vhd at path %s : %s", cfDevVHD, err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Add-VMHardDiskDrive -VMName %s "+
		"-Path %s", vmName, cfDevVHD))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("adding vhd %s : %s", cfDevVHD, err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Set-VMFirmware "+
		"-VMName %s "+
		"-EnableSecureBoot Off "+
		"-FirstBootDevice $cdrom",
		vmName))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("setting firmware : %s", err)
	}
	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Set-VMComPort "+
		"-VMName %s "+
		"-number 1 "+
		"-Path \\\\.\\pipe\\cfdev-com",
		vmName))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("setting com port : %s", err)
	}

	return nil
}

func addVhdDrive(isoPath string, vmName string) error {
	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Add-VMDvdDrive -VMName %s -Path %s", vmName, isoPath))
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (h *HyperV) Start(vmName string) error {
	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s", vmName))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("start-vm: %s : %s", err, string(output))
	}

	return nil
}

func (h *HyperV) Stop(vmName string) error {
	cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev*")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("getting vms: %s", err)
	}

	if string(output) == "" {
		return nil
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Stop-VM -Name %s -Turnoff", vmName))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("stopping vm: %s", err)
	}

	return nil
}

func (h *HyperV) Destroy(vmName string) error {
	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("removing vm: %s", err)
	}

	return nil
}

func (h *HyperV) IsRunning() (bool, error) {
	//TODO implement this
	return false, nil
}
