package kvm

import (
	"code.cloudfoundry.org/cfdev/driver"
	"os/exec"
)

// consider moving to more native go implementation
// rather than shelling out
func (d *KVM) setupNetworking(tapDevice, bridge string) {
	output, err := exec.Command("ip", "tuntap", "add", "dev", tapDevice, "mode", "tap").CombinedOutput()
	if err != nil {
		d.UI.Say("[WARNING] adding tap device failed: %s: %s", err, output)
	}

	output, err = exec.Command("ip", "link", "set", tapDevice, "master", bridge).CombinedOutput()
	if err != nil {
		d.UI.Say("[WARNING] adding master failed: %s: %s", err, output)
	}

	output, err = exec.Command("ip", "link", "set", "dev", bridge, "up").CombinedOutput()
	if err != nil {
		d.UI.Say("[WARNING] turning virtual bridge device failed: %s: %s", err, output)
	}

	output, err = exec.Command("ip", "link", "set", "dev", tapDevice, "up").CombinedOutput()
	if err != nil {
		d.UI.Say("[WARNING] turning tap device failed: %s: %s", err, output)
	}
}

func (d *KVM) setupRoutes(ip string) {
	output, err := exec.Command("ip", "route", "add", driver.ContainerSubnet, "via", ip).CombinedOutput()
	if err != nil {
		d.UI.Say("[WARNING] adding container routes via %s failed: %s: %s", ip, err, output)
	}
}

func (d *KVM) teardownRoutes() {
	exec.Command("ip", "route", "flush", driver.ContainerSubnet).Run()
}

func (d *KVM) teardownNetworking(tapDevice string) {
	exec.Command("ip", "link", "set", "dev", tapDevice, "down").Run()
	exec.Command("ip", "link", "del", "dev", tapDevice).Run()
}