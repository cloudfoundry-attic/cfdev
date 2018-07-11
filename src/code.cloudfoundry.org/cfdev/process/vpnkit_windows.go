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

	fmt.Println("VCP A")

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

	fmt.Println("VCP B")

	resolvConfPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	if fileExists(resolvConfPath) {
		os.RemoveAll(resolvConfPath)
	}

	err = ioutil.WriteFile(resolvConfPath, []byte(dnsFile), 0600)
	if err != nil {
		return err
	}

	fmt.Println("VCP C")

	cmd := exec.Command("powershell.exe", "-Command", "get-dnsclient | select-object -expandproperty connectionspecificsuffix")
	dhcp, err := cmd.Output()
	if err != nil {
		return err
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

	fmt.Println("VCP D")

	dhcpJsonPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")
	if fileExists(dhcpJsonPath) {
		os.RemoveAll(dhcpJsonPath)
	}

	file, err := os.Create(dhcpJsonPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&output)
}

func (v *VpnKit) Start() error {

	fmt.Println("CHECK A")

	if err := v.Setup(); err != nil {
		return errors.SafeWrap(err, "Failed to setup VPNKit")
	}

	fmt.Println("CHECK B")

	cmd := exec.Command("powershell.exe", "-Command", "((Get-VM -Name cfdev).Id).Guid")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd.Wait()
	vmGuid := strings.TrimSpace(string(output))
	fmt.Println("CHECK C")
	fmt.Println("VM GUID: " + string(vmGuid))

	if err := v.Launchd.AddDaemon(v.daemonSpec(vmGuid)); err != nil {
		return errors.SafeWrap(err, "install vpnkit")
	}

	fmt.Println("CHECK D")

	if err := v.Launchd.Start(v.daemonSpec(vmGuid)); err != nil {
		return errors.SafeWrap(err, "start vpnkit")
	}

	fmt.Println("CHECK E")
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
	/*
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
	*/

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

func (v *VpnKit) daemonSpec(vmGuid string) launchd.DaemonSpec {
	dnsPath := filepath.Join(v.Config.CFDevHome, "resolv.conf")
	dhcpPath := filepath.Join(v.Config.CFDevHome, "dhcp.json")

	return launchd.DaemonSpec{
		Label:     VpnKitLabel,
		Program:   path.Join(v.Config.CacheDir, "vpnkit.exe"),
		CfDevHome: v.Config.CFDevHome,
		ProgramArguments: []string{
			fmt.Sprintf("--ethernet hyperv-connect://%s/7207f451-2ca3-4b88-8d01-820a21d78293", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa", vmGuid),
			fmt.Sprintf("--port hyperv-connect://%s/e3ae8f06-8c25-47fb-b6ed-c20702bcef5e", vmGuid),
			fmt.Sprintf("--dns %s", dnsPath),
			fmt.Sprintf("--dhcp %s", dhcpPath),
			"--diagnostics \\\\.\\pipe\\cfdevVpnKitDiagnostics",
			"--listen-backlog 32",
			"--lowest-ip 169.254.82.3",
			"--highest-ip 169.254.82.255",
			"--host-ip 169.254.82.2",
			"--gateway-ip 169.254.82.1",
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
