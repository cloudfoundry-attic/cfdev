package process

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/errors"
	"bufio"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

const ethernetGUID = "7207f451-2ca3-4b88-8d01-820a21d78293"
const portGUID = "cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa"
const forwarderGUID = "e3ae8f06-8c25-47fb-b6ed-c20702bcef5e"

func (v *VpnKit) Start() error {
	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to Setup VPNKit")
	}

	cmd := exec.Command("powershell.exe", "-Command", "((Get-VM -Name cfdev).Id).Guid")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("get vm name: %s", err)
	}

	cmd.Wait()
	vmGuid := strings.TrimSpace(string(output))

	if err := v.DaemonRunner.AddDaemon(v.daemonSpec(vmGuid)); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}

	if err := v.DaemonRunner.Start(VpnKitLabel); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}

	return nil
}

func (v *VpnKit) Destroy() error {
	v.DaemonRunner.RemoveDaemon(VpnKitLabel)
	registryDeleteCmd := `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
		`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
		`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`
	if err := exec.Command("powershell.exe", "-Command", registryDeleteCmd).Run(); err != nil {
		return fmt.Errorf("failed to remove service registries: %s", err)
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
			fmt.Sprintf("--ethernet hyperv-connect://%s/%s", vmGuid, ethernetGUID),
			fmt.Sprintf("--port hyperv-connect://%s/%s", vmGuid, portGUID),
			fmt.Sprintf("--port hyperv-connect://%s/%s", vmGuid, forwarderGUID),
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
	if err := v.registerGUID(ethernetGUID, "Ethernet"); err != nil {
		return err
	}
	if err := v.registerGUID(portGUID, "Port"); err != nil {
		return err
	}
	return v.registerGUID(forwarderGUID, "Forwarder")
}

func (v *VpnKit) registerGUID(guid, name string) error {
	command := exec.Command(
		"powershell.exe", "-Command",
		fmt.Sprintf(`$ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name %s;
             $ethService.SetValue("ElementName", "CF Dev VPNkit %s Service" )`, guid, name))

	return command.Run()
}

func (v *VpnKit) writeResolvConf() error{
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

	return ioutil.WriteFile(resolvConfPath, []byte(dnsFile), 0600)
}

func (v *VpnKit) writeDHCPJSON() error{
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

	scanner := bufio.NewScanner(bytes.NewReader(dhcp))
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
