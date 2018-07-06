package process

import (
	"code.cloudfoundry.org/cfdev/config"
	"os/exec"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"os"
	"encoding/json"
	"path/filepath"
)

type VpnKit struct {
	Config config.Config
	Launchd Launchd

}

func (v *VpnKit) Setup() error {
	err := v.generateServiceGUIDs()
	if err != nil {
		return err
	}

	dns, err := exec.Command("powershell.exe", "-Command", "get-dnsclientserveraddress -family ipv4 | select-object -expandproperty serveraddresses").Output()
	if err != nil {
		return err
	}

	dnsFile := ""
	scanner := bufio.NewScanner(bytes.NewReader(dns))
	for scanner.Scan() {
		line := scanner.Text()
		dnsFile += fmt.Sprintf("nameserver %s\r\n", line)
	}

	err = ioutil.WriteFile(filepath.Join(v.Config.CFDevHome, "resolv.conf"), []byte(dnsFile), 0600)
	if err != nil {
		return err
	}

	dhcp, err := exec.Command("powershell.exe", "-Command", "get-dnsclient | select-object -expandproperty connectionspecificsuffix").Output()
	if err != nil {
		return err
	}

	var output struct {
		SearchDomains []string `json:"searchDomains"`
		DomainName string `json:"domainName"`
	}

	scanner = bufio.NewScanner(bytes.NewReader(dhcp))
	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			output.SearchDomains = append(output.SearchDomains, line)
		}

		if len(output.SearchDomains) > 0 {
			output.DomainName = output.SearchDomains[len(output.SearchDomains) - 1]
		}
	}

	file, err := os.Create(filepath.Join(v.Config.CFDevHome, "dhcp.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&output)
}

func (v *VpnKit) Start() error {
	return nil
}

func (v *VpnKit) Stop() {
}

func (v *VpnKit) Watch(exit chan string) {

}

func (v *VpnKit) generateServiceGUIDs() error {
	for _, serviceName := range []string{"CF Dev VPNkit Ethernet Service", "CF Dev VPNkit Port Service", "CF Dev VPNkit Forwarder Service"} {
		command := exec.Command(
			"powershell.exe", "-Command",
			fmt.Sprintf(`$guid=[guid]::newguid().Guid;
			  $ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name $guid;
             $ethService.SetValue("ElementName", "%s")
             `, serviceName),
		)

		if err := command.Run(); err != nil {
			return err
		}
	}

	return nil
}