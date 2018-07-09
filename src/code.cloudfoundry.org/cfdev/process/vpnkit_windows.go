package process

import (
	"bufio"
	"bytes"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdevd/launchd"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"code.cloudfoundry.org/cfdev/errors"
)

type VpnKit struct {
	Config  config.Config
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

	file, err := os.Create(filepath.Join(v.Config.CFDevHome, "dhcp.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&output)
}

func (v *VpnKit) Start() error {
	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to setup VPNKit")
	}

	cmd := exec.Command("powershell.exe", "-Command", "((Get-VM -Name cfdev).Id).Guid")
	vmGuid, err := cmd.Output()
	if err != nil {
		return err
	}

	if err := v.Launchd.AddDaemon(v.daemonSpec(string(vmGuid))); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}

	//if err := v.Launchd.Start(VpnKitLabel); err != nil {
	//	return errors.SafeWrap(err, "start vpnkit")
	//}
	//attempt := 0
	//for {
	//	conn, err := net.Dial("unix", filepath.Join(v.Config.VpnKitStateDir, "vpnkit_eth.sock"))
	//	if err == nil {
	//		conn.Close()
	//		return nil
	//	} else if attempt >= retries {
	//		return errors.SafeWrap(err, "conenct to vpnkit")
	//	} else {
	//		time.Sleep(time.Second)
	//		attempt++
	//	}
	//}

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

func (v *VpnKit) daemonSpec(vmGuid string) launchd.DaemonSpec {
	dnsPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	dhcpPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")

	return launchd.DaemonSpec{
		Label:   VpnKitLabel,
		Program: path.Join(v.Config.CacheDir, "vpnkit.exe"),
		ProgramArguments: []string{
			fmt.Sprintf("--ethernet hyperv-connect://%s/'7207f451-2ca3-4b88-8d01-820a21d78293'", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/'cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa'", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/'e3ae8f06-8c25-47fb-b6ed-c20702bcef5e'", vmGuid),
			fmt.Sprintf("--dns %s", dnsPath),
			fmt.Sprintf("--dhcp %s", dhcpPath),
			"--diagnostics '\\\\.\\pipe\\cfdevVpnKitDiagnostics'",
			"--listen-backlog 32",
			"--lowest-ip 169.254.82.3",
			"--highest-ip 169.254.82.255",
			"--host-ip 169.254.82.2",
			fmt.Sprintf("--gateway-ip 169.254.82.1 2>&1 > %s", filepath.Join(v.Config.CFDevHome, "vpnkit.log")),
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(v.Config.CFDevHome, "vpnkit.stdout.log"),
		StderrPath: path.Join(v.Config.CFDevHome, "vpnkit.stderr.log"),
	}
}
