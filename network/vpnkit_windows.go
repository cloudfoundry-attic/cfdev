package network

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bufio"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/errors"
	"encoding/json"
	"io/ioutil"
)

func (v *VpnKit) Start() error {
	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to Setup VPNKit")
	}

	output, err := v.Powershell.Output("((Get-VM -Name cfdev).Id).Guid")
	if err != nil {
		return fmt.Errorf("get vm name: %s", err)
	}

	vmGuid := strings.TrimSpace(output)

	if err := v.DaemonRunner.AddDaemon(v.daemonSpec(vmGuid)); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}

	if err := v.DaemonRunner.Start(v.Label); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}

	return nil
}

func (v *VpnKit) Destroy() error {
	v.DaemonRunner.RemoveDaemon(v.Label)

	registryDeleteCmd := `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
		`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
		`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`

	_, err := v.Powershell.Output(registryDeleteCmd)
	if err != nil {
		return fmt.Errorf("failed to remove service registries: %s", err)
	}

	return nil
}

func (v *VpnKit) daemonSpec(vmGuid string) daemon.DaemonSpec {
	dnsPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	dhcpPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")

	return daemon.DaemonSpec{
		Label:   v.Label,
		Program: path.Join(v.Config.CacheDir, "vpnkit.exe"),
		ProgramArguments: []string{
			fmt.Sprintf("--ethernet hyperv-connect://%s/%s", vmGuid, v.EthernetGUID),
			fmt.Sprintf("--port hyperv-connect://%s/%s", vmGuid, v.PortGUID),
			fmt.Sprintf("--port hyperv-connect://%s/%s", vmGuid, v.ForwarderGUID),
			fmt.Sprintf("--dns %s", dnsPath),
			fmt.Sprintf("--dhcp %s", dhcpPath),
			"--http", path.Join(v.Config.VpnKitStateDir, "http_proxy.json"),
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

func (v *VpnKit) Setup() error {
	if err := v.registerServiceGUIDs(); err != nil {
		return fmt.Errorf("generating service guids: %s", err)
	}

	if err := v.writeHttpConfig(); err != nil {
		return err
	}

	if err := v.writeResolvConf(); err != nil {
		return fmt.Errorf("writing resolv.conf: %s", err)
	}

	if err := v.writeDHCPJSON(); err != nil {
		return fmt.Errorf("writing dhcp.json: %s", err)
	}
	return nil
}

func (v *VpnKit) registerServiceGUIDs() error {
	if err := v.registerGUID(v.EthernetGUID, "Ethernet"); err != nil {
		return err
	}
	if err := v.registerGUID(v.PortGUID, "Port"); err != nil {
		return err
	}
	return v.registerGUID(v.ForwarderGUID, "Forwarder")
}

func (v *VpnKit) registerGUID(guid, name string) error {
	command := fmt.Sprintf(`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name %s;
             $ethService.SetValue("ElementName", "CF Dev VPNkit %s Service" )`, guid, name)

	_, err := v.Powershell.Output(command)
	return err
}

func (v *VpnKit) writeResolvConf() error {
	command := "get-dnsclientserveraddress -family ipv4 | select-object -expandproperty serveraddresses"
	dns, err := v.Powershell.Output(command)
	if err != nil {
		return fmt.Errorf("getting dns client server addresses: %s", err)
	}

	dnsFile := ""
	scanner := bufio.NewScanner(strings.NewReader(dns))
	for scanner.Scan() {
		line := scanner.Text()
		dnsFile += fmt.Sprintf("nameserver %s\r\n", line)
	}

	resolvConfPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	if fileExists(resolvConfPath) {
		os.RemoveAll(resolvConfPath)
	}

	return ioutil.WriteFile(resolvConfPath, []byte(dnsFile), 0600)
}

func (v *VpnKit) writeDHCPJSON() error {
	command := "get-dnsclient | select-object -expandproperty connectionspecificsuffix"
	dhcp, err := v.Powershell.Output(command)
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
