package process

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/errors"
)

type VpnKit struct {
	Config  config.Config
	Launchd Launchd
}

func (v *VpnKit) Setup() error {
	err := v.generateServiceGUIDs()
	if err != nil {
		return fmt.Errorf("generating service guids: %s", err)
	}

	dns, err := exec.Command("powershell.exe", "-Command", "get-dnsclientserveraddress -family ipv4 | select-object -expandproperty serveraddresses").Output()
	if err != nil {
		return fmt.Errorf("getting dns client server addresses: %s", err)
	}

	dnsFile := ""
	scanner := bufio.NewScanner(bytes.NewReader(dns))
	for scanner.Scan() {
		line := scanner.Text()
		dnsFile += fmt.Sprintf("nameserver %s\r\n", line)
	}

	resolvConfPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	if fileExists(resolvConfPath) {
		os.RemoveAll(resolvConfPath)
	}

	err = ioutil.WriteFile(resolvConfPath, []byte(dnsFile), 0600)
	if err != nil {
		return fmt.Errorf("writing resolv.conf: %s", err)
	}

	cmd := exec.Command("powershell.exe", "-Command", "get-dnsclient | select-object -expandproperty connectionspecificsuffix")
	dhcp, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("get dns client: %s", err)
	}

	cmd.Wait()

	var output struct {
		SearchDomains []string `json:"searchDomains"`
		DomainName    string   `json:"domainName"`
	}

	scanner = bufio.NewScanner(bytes.NewReader(dhcp))
	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			output.SearchDomains = append(output.SearchDomains, line)
		}

		if len(output.SearchDomains) > 0 {
			output.DomainName = output.SearchDomains[len(output.SearchDomains)-1]
		}
	}

	dhcpJsonPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")
	if fileExists(dhcpJsonPath) {
		os.RemoveAll(dhcpJsonPath)
	}

	file, err := os.Create(dhcpJsonPath)
	if err != nil {
		return fmt.Errorf("creating dhcp.json: %s", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&output)
}

func (v *VpnKit) Start() error {
	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to setup VPNKit")
	}

	cmd := exec.Command("powershell.exe", "-Command", "((Get-VM -Name cfdev).Id).Guid")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("get vm name: %s", err)
	}

	cmd.Wait()
	vmGuid := strings.TrimSpace(string(output))

	if err := v.Launchd.AddDaemon(v.daemonSpec(vmGuid)); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}

	if err := v.Launchd.Start(VpnKitLabel); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}

	return nil
}

func (v *VpnKit) Destroy() error {
	v.Launchd.RemoveDaemon(VpnKitLabel)
	registryDeleteCmd := `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
		`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
		`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`
	if err := exec.Command("powershell.exe", "-Command", registryDeleteCmd).Run(); err != nil {
		return fmt.Errorf("failed to remove service registries: %s", err)
	}
	return nil
}

func (v *VpnKit) Watch(exit chan string) {

}

func (v *VpnKit) generateServiceGUIDs() error {
	command := exec.Command(
		"powershell.exe", "-Command",
		`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name 7207f451-2ca3-4b88-8d01-820a21d78293;
             $ethService.SetValue("ElementName", "CF Dev VPNkit Ethernet Service" )`)

	if err := command.Run(); err != nil {
		return err
	}

	command = exec.Command(
		"powershell.exe", "-Command",
		`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa;
             $ethService.SetValue("ElementName", "CF Dev VPNkit Port Service" )`)

	if err := command.Run(); err != nil {
		return err
	}

	command = exec.Command(
		"powershell.exe", "-Command",
		`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name e3ae8f06-8c25-47fb-b6ed-c20702bcef5e;
             $ethService.SetValue("ElementName", "CF Dev VPNkit Forwarder Service" )`)

	if err := command.Run(); err != nil {
		return err
	}

	return nil
}

func (v *VpnKit) daemonSpec(vmGuid string) daemon.DaemonSpec {
	dnsPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	dhcpPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")

	return daemon.DaemonSpec{
		Label:   VpnKitLabel,
		Program: path.Join(v.Config.CacheDir, "vpnkit.exe"),
		ProgramArguments: []string{
			fmt.Sprintf("--ethernet hyperv-connect://%s/7207f451-2ca3-4b88-8d01-820a21d78293", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/e3ae8f06-8c25-47fb-b6ed-c20702bcef5e", vmGuid),
			fmt.Sprintf("--dns %s", dnsPath),
			fmt.Sprintf("--dhcp %s", dhcpPath),
			"--diagnostics \\\\.\\pipe\\cfdevVpnKitDiagnostics",
			"--listen-backlog 32",
			"--lowest-ip 192.168.65.3",
			"--highest-ip 192.168.65.255",
			"--host-ip 192.168.65.2",
			"--gateway-ip 192.168.65.1",
			"--host-names host.cfdev.sh",
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(v.Config.CFDevHome, "vpnkit.stdout.log"),
		StderrPath: path.Join(v.Config.CFDevHome, "vpnkit.stderr.log"),
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}
