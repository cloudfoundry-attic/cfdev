package host

import (
	safeerr "code.cloudfoundry.org/cfdev/errors"
	"errors"
	"fmt"
	"strings"
)

const admin_role = "[Security.Principal.WindowsBuiltInRole]::Administrator"
const current_user = "New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())"

func (h *Host) CheckRequirements() error {
	if err := h.hasAdminPrivileged(); err != nil {
		return err
	}
	return h.hypervEnabled()
}


func (h *Host) hasAdminPrivileged() error {
	command := fmt.Sprintf("(%s).IsInRole(%s)", current_user, admin_role)
	output, err := h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("checking for admin privileges: %s", err)
	}

	if strings.TrimSpace(output) == "True" {
		return nil
	}

	return safeerr.SafeWrap(errors.New("You must run cf dev with an admin privileged powershell"),"Running without admin privileges")
}

const hyperv_feature="Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online"
const hyperv_disabled_error=`You must first enable Hyper-V on your machine before you run CF Dev. Please use the following tutorial to enable this functionality on your machine

https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v`

func (h *Host) hypervEnabled() error {
	command := fmt.Sprintf("(%s).State", hyperv_feature)
	output, err := h.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("checking whether hyperv is enabled: %s", err)
	}

	if strings.TrimSpace(output) == "Enabled" {
		return nil
	}

	return safeerr.SafeWrap(errors.New(hyperv_disabled_error),"Hyper-V disabled")
}
