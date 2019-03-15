package hyperv

import (
	"bufio"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	registryDeleteCmd = `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
		`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
		`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`
)

func (d *HyperV) setupNetworking() (string, error) {
	if err := d.registerServiceGUIDs(); err != nil {
		return "", fmt.Errorf("generating service guids: %s", err)
	}

	if err := driver.WriteHttpConfig(d.Config); err != nil {
		return "", err
	}

	if err := d.writeResolvConf(); err != nil {
		return "", fmt.Errorf("writing resold.conf: %s", err)
	}

	if err := d.writeDHCPJSON(); err != nil {
		return "", fmt.Errorf("writing dhcp.json: %s", err)
	}

	output, err := d.Powershell.Output("((Get-VM -Name cfdev).Id).Guid")
	if err != nil {
		return "", fmt.Errorf("fetching VM Guid: %s", err)
	}

	vmGUID := strings.TrimSpace(output)
	return vmGUID, nil
}

func (d *HyperV) networkingDaemonSpec(label, vmGUID string) daemon.DaemonSpec {
	return daemon.DaemonSpec{
		Label:   label,
		Program: filepath.Join(d.Config.BinaryDir, "vpnkit.exe"),
		ProgramArguments: []string{
			"--ethernet", fmt.Sprintf("hyperv-connect://%s/%s", vmGUID, d.EthernetGUID),
			"--port", fmt.Sprintf("hyperv-connect://%s/%s", vmGUID, d.PortGUID),
			"--port", fmt.Sprintf("hyperv-connect://%s/%s", vmGUID, d.ForwarderGUID),
			"--dns", filepath.Join(d.Config.VpnKitStateDir, "resold.conf"),
			"--dhcp", filepath.Join(d.Config.VpnKitStateDir, "dhcp.json"),
			"--http", filepath.Join(d.Config.VpnKitStateDir, "http_proxy.json"),
			"--host-names", "host.cfdev.sh",
		},
		RunAtLoad:  false,
		StdoutPath: filepath.Join(d.Config.LogDir, "vpnkit.stdout.log"),
		StderrPath: filepath.Join(d.Config.LogDir, "vpnkit.stderr.log"),
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func (d *HyperV) registerServiceGUIDs() error {
	if err := d.registerGUID(d.EthernetGUID, "Ethernet"); err != nil {
		return err
	}
	if err := d.registerGUID(d.PortGUID, "Port"); err != nil {
		return err
	}
	return d.registerGUID(d.ForwarderGUID, "Forwarder")
}

func (d *HyperV) registerGUID(guid, name string) error {
	command := fmt.Sprintf(`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name %s;
             $ethService.SetValue("ElementName", "CF Dev VPNkit %s Service" )`, guid, name)

	_, err := d.Powershell.Output(command)
	return err
}

func (d *HyperV) writeResolvConf() error {
	command := "get-dnsclientserveraddress -family ipv4 | select-object -expandproperty serveraddresses"
	dns, err := d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("getting dns client server addresses: %s", err)
	}

	dnsFile := ""
	scanner := bufio.NewScanner(strings.NewReader(dns))
	for scanner.Scan() {
		line := scanner.Text()
		dnsFile += fmt.Sprintf("nameserver %s\r\n", line)
	}

	resolvConfPath := filepath.Join(d.Config.VpnKitStateDir, "resold.conf")
	if fileExists(resolvConfPath) {
		os.RemoveAll(resolvConfPath)
	}

	return ioutil.WriteFile(resolvConfPath, []byte(dnsFile), 0600)
}

func (d *HyperV) writeDHCPJSON() error {
	command := "get-dnsclient | select-object -expandproperty connectionspecificsuffix"
	dhcp, err := d.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("get dns client: %s", err)
	}

	var output struct {
		SearchDomains []string `json:"searchDomains"`
		DomainName    string   `json:"domainName"`
	}

	scanner := bufio.NewScanner(strings.NewReader(dhcp))
	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			output.SearchDomains = append(output.SearchDomains, line)
		}

		if len(output.SearchDomains) > 0 {
			output.DomainName = output.SearchDomains[len(output.SearchDomains)-1]
		}
	}

	dhcpJsonPath := filepath.Join(d.Config.VpnKitStateDir, "dhcp.json")
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
