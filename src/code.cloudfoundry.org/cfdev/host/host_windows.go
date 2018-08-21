package host

import (
	"os/exec"
	"fmt"
	"strings"
	safeerr "code.cloudfoundry.org/cfdev/errors"
	"errors"
)

const admin_role = "[Security.Principal.WindowsBuiltInRole]::Administrator"
const current_user = "New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())"

func (*Host) CheckRequirements() error {
	if err := hasAdminPrivileged(); err != nil {
		return err
	}
	return hypervEnabled()
}


func hasAdminPrivileged() error {
	cmd := exec.Command("powershell.exe", "-Command",
		fmt.Sprintf("(%s).IsInRole(%s)", current_user, admin_role))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("checking for admin privileges: %s", err)
	}
	if strings.TrimSpace(string(output)) == "True" {
		return nil
	}
	return safeerr.SafeWrap(errors.New("You must run cf dev with an admin privileged powershell"),"Running without admin privileges")
}

const hyperv_feature="Get-WindowsOptionalFeature -FeatureName Microsoft-Hyper-V-All -Online"
const hyperv_disabled_error=`You must first enable Hyper-V on your machine before you run CF Dev. Please use the following tutorial to enable this functionality on your machine

https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v`
func hypervEnabled() error {
	cmd := exec.Command("powershell.exe", "-Command",
		fmt.Sprintf("(%s).State", hyperv_feature))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("checking whether hyperv is enabled: %s", err)
	}
	if strings.TrimSpace(string(output)) == "Enabled" {
		return nil
	}
	return safeerr.SafeWrap(errors.New(hyperv_disabled_error),"Hyper-V disabled")
}