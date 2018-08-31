package hypervisor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"code.cloudfoundry.org/cfdev/config"
)

type HyperV struct {
	Config config.Config
}

func (h *HyperV) CreateVM(vm VM) error {
	var cfdevEfiIso = filepath.Join(h.Config.CacheDir, "cfdev-efi.iso")
	if vm.DepsIso == "" {
		vm.DepsIso = filepath.Join(h.Config.CacheDir, "cf-deps.iso")
	}
	var cfDevVHD = filepath.Join(h.Config.CFDevHome, "cfdev.vhd")

	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vm.Name))
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
		vm.Name))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("setting vm properites (memoryMB:%d, cpus:%d): %s", vm.MemoryMB, vm.CPUs, err)
	}

	err = addVhdDrive(cfdevEfiIso, vm.Name)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", cfdevEfiIso, err)
	}

	err = addVhdDrive(vm.DepsIso, vm.Name)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", vm.DepsIso, err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VMNetworkAdapter "+
		"-VMName %s "+
		"-Name 'Network Adapter'",
		vm.Name))
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
		"-Path %s", vm.Name, cfDevVHD))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("adding vhd %s : %s", cfDevVHD, err)
	}

	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Set-VMFirmware "+
		"-VMName %s "+
		"-EnableSecureBoot Off "+
		"-FirstBootDevice $cdrom",
		vm.Name))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("setting firmware : %s", err)
	}
	cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Set-VMComPort "+
		"-VMName %s "+
		"-number 1 "+
		"-Path \\\\.\\pipe\\cfdev-com",
		vm.Name))
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

func (h *HyperV) exists(vmName string) (bool, error) {
	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s*", vmName))
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("getting vms: %s", err)
	}

	return string(output) != "", nil
}
func (h *HyperV) Start(vmName string) error {
	if exists, err := h.exists(vmName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("hyperv vm with name %s does not exist", vmName)
	}

	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s", vmName))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("start-vm: %s : %s", err, string(output))
	}

	return nil
}

func (h *HyperV) Stop(vmName string) error {
	if exists, err := h.exists(vmName); err != nil {
		return err
	} else if !exists {
		return nil
	}

	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Stop-VM -Name %s -Turnoff", vmName))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stopping vm: %s", err)
	}

	return nil
}

func (h *HyperV) Destroy(vmName string) error {
	if exists, err := h.exists(vmName); err != nil {
		return err
	} else if !exists {
		return nil
	}

	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("removing vm: %s", err)
	}

	return nil
}

func (h *HyperV) IsRunning(vmName string) (bool, error) {
	if exists, err := h.exists(vmName); err != nil || !exists {
		return false, err
	}
	cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	if strings.Contains(string(output), "Running") {
		return true, nil
	}
	return false, nil
}
