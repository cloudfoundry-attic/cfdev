package hypervisor

import (
	"code.cloudfoundry.org/cfdev/runner"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"code.cloudfoundry.org/cfdev/config"
)

type HyperV struct {
	Config     config.Config
	Powershell runner.Powershell
}

func (h *HyperV) CreateVM(vm VM) error {
	var cfdevEfiIso = filepath.Join(h.Config.CacheDir, "cfdev-efi.iso")
	if vm.DepsIso == "" {
		vm.DepsIso = filepath.Join(h.Config.CacheDir, "cf-deps.iso")
	}
	var cfDevVHD = filepath.Join(h.Config.CFDevHome, "cfdev.vhd")

	command := fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vm.Name)
	_, err := h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("creating new vm: %s", err)
	}

	command = fmt.Sprintf("Set-VM -Name %s "+
		"-AutomaticStartAction Nothing "+
		"-AutomaticStopAction ShutDown "+
		"-CheckpointType Disabled "+
		fmt.Sprintf("-MemoryStartupBytes %dMB ", vm.MemoryMB)+
		"-StaticMemory "+
		fmt.Sprintf("-ProcessorCount %d", vm.CPUs),
		vm.Name)
	_, err = h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting vm properites (memoryMB:%d, cpus:%d): %s", vm.MemoryMB, vm.CPUs, err)
	}

	err = h.addVhdDrive(cfdevEfiIso, vm.Name)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", cfdevEfiIso, err)
	}

	err = h.addVhdDrive(vm.DepsIso, vm.Name)
	if err != nil {
		return fmt.Errorf("adding dvd drive %s: %s", vm.DepsIso, err)
	}

	command = fmt.Sprintf("(Get-VMNetworkAdapter -VMName * | Where-Object -FilterScript {$_.VMName -eq '%s'}).Name", vm.Name)
	output, err := h.Powershell.Output(command)
	if err == nil {
		if output != "" {
			adapterNames := strings.Split(output, "\n")
			for _, name := range adapterNames {
				command = fmt.Sprintf("Remove-VMNetworkAdapter "+
					"-VMName %s "+
					"-Name '%s'",
					vm.Name, strings.TrimSpace(name))

				_, err = h.Powershell.Output(command)
				if err != nil {
					fmt.Printf("failed to remove netowork adapter: %s", err)
				}
			}
		}
	}

	if _, err := os.Stat(cfDevVHD); err == nil {
		err := os.RemoveAll(cfDevVHD)
		if err != nil {
			return fmt.Errorf("removing any vhds: %s", err)
		}
	}

	command = fmt.Sprintf(`New-VHD -Path "%s" `+
		"-SizeBytes '200000000000' "+
		"-Dynamic", cfDevVHD)
	_, err = h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("creating new vhd at path %s : %s", cfDevVHD, err)
	}

	command = fmt.Sprintf("Add-VMHardDiskDrive -VMName %s "+
		`-Path "%s"`, vm.Name, cfDevVHD)
	_, err = h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("adding vhd %s : %s", cfDevVHD, err)
	}

	command = fmt.Sprintf("Set-VMFirmware "+
		"-VMName %s "+
		"-EnableSecureBoot Off "+
		"-FirstBootDevice $cdrom",
		vm.Name)
	_, err = h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting firmware : %s", err)
	}

	command = fmt.Sprintf("Set-VMComPort "+
		"-VMName %s "+
		"-number 1 "+
		"-Path \\\\.\\pipe\\cfdev-com",
		vm.Name)
	_, err = h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("setting com port : %s", err)
	}

	return nil
}

func (h *HyperV) addVhdDrive(isoPath string, vmName string) error {
	command := fmt.Sprintf(`Add-VMDvdDrive -VMName %s -Path "%s"`, vmName, isoPath)
	_, err := h.Powershell.Output(command)
	if err != nil {
		return err
	}

	return nil
}

func (h *HyperV) exists(vmName string) (bool, error) {
	command := fmt.Sprintf("Get-VM -Name %s*", vmName)
	output, err := h.Powershell.Output(command)
	if err != nil {
		return false, fmt.Errorf("getting vms: %s", err)
	}

	return output != "", nil
}
func (h *HyperV) Start(vmName string) error {
	if exists, err := h.exists(vmName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("hyperv vm with name %s does not exist", vmName)
	}

	command := fmt.Sprintf("Start-VM -Name %s", vmName)
	if _, err := h.Powershell.Output(command); err != nil {
		return fmt.Errorf("start-vm: %s", err)
	}

	return nil
}

func (h *HyperV) Stop(vmName string) error {
	if exists, err := h.exists(vmName); err != nil {
		return err
	} else if !exists {
		return nil
	}

	command := fmt.Sprintf("Stop-VM -Name %s -Turnoff", vmName)
	if _, err := h.Powershell.Output(command); err != nil {
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

	command := fmt.Sprintf("Remove-VM -Name %s -Force", vmName)
	if _, err := h.Powershell.Output(command); err != nil {
		return fmt.Errorf("removing vm: %s", err)
	}

	return nil
}

func (h *HyperV) IsRunning(vmName string) (bool, error) {
	if exists, err := h.exists(vmName); err != nil || !exists {
		return false, err
	}

	command :=  fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName)
	output, err := h.Powershell.Output(command)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(output), "Running") {
		return true, nil
	}

	return false, nil
}
