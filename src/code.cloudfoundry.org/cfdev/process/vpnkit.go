package process

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
)

type VpnKit struct {
	HomeDir  string
	CacheDir string
	StateDir string
}

func (v *VpnKit) Command() *exec.Cmd {
	cmd := exec.Command(path.Join(v.CacheDir, "vpnkit"),
		"--ethernet",
		path.Join(v.HomeDir, "vpnkit_eth.sock"),
		"--port",
		path.Join(v.HomeDir, "vpnkit_port.sock"),
		"--vsock-path",
		path.Join(v.StateDir, "connect"),
		"--http",
		path.Join(v.HomeDir, "http_proxy.json"))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}

func (v *VpnKit) SetupVPNKit(homeDir string) {
	vpnkitEthPath := path.Join(homeDir, "vpnkit_eth.sock")
	vpnkitPortPath := path.Join(homeDir, "vpnkit_port.sock")
	httpProxyPath := path.Join(homeDir, "http_proxy.json")

	if _, err := os.Stat(vpnkitEthPath); err != nil {
		err := ioutil.WriteFile(vpnkitEthPath, []byte(""), 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to setup VPNKit dependencies %v\n", err)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(vpnkitPortPath); err != nil {
		err = ioutil.WriteFile(vpnkitPortPath, []byte(""), 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to setup VPNKit dependencies %v\n", err)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(httpProxyPath); err != nil {
		//httpProxyConfig := []byte("{\"http\":\"localhost:8989\"}")
		httpProxyConfig := []byte("{}")
		err = ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to setup VPNKit dependencies %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("writing %s to %s", string(httpProxyConfig), httpProxyPath)
	}
}
