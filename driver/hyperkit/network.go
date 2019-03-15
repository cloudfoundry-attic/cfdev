package hyperkit

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/driver"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

func (d *Hyperkit) waitForNetworking() error {
	var (
		ticker  = time.NewTicker(time.Second)
		timeout = time.After(30 * time.Second)
		err     error
	)

	for {
		select {
		case <-ticker.C:
			var conn net.Conn
			conn, err = net.Dial("unix", filepath.Join(d.Config.VpnKitStateDir, "vpnkit_eth.sock"))
			if err == nil {
				conn.Close()
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timed out connecting to vpnkit: %s", err)
		}
	}
}

func (d *Hyperkit) networkingDaemonSpec() daemon.DaemonSpec {
	return daemon.DaemonSpec{
		Label:       driver.VpnKitLabel,
		Program:     path.Join(d.Config.BinaryDir, "vpnkit"),
		SessionType: "Background",
		ProgramArguments: []string{
			path.Join(d.Config.CacheDir, "vpnkit"),
			"--ethernet", path.Join(d.Config.VpnKitStateDir, "vpnkit_eth.sock"),
			"--port", path.Join(d.Config.VpnKitStateDir, "vpnkit_port.sock"),
			"--vsock-path", path.Join(d.Config.StateLinuxkit, "connect"),
			"--http", path.Join(d.Config.VpnKitStateDir, "http_proxy.json"),
			"--host-names", "host.cfded.sh",
		},
		RunAtLoad:  false,
		StdoutPath: path.Join(d.Config.LogDir, "vpnkit.stdout.log"),
		StderrPath: path.Join(d.Config.LogDir, "vpnkit.stderr.log"),
	}
}

func (d *Hyperkit) installCFDevDaemon() error {
	var (
		executablePath = filepath.Join(d.Config.CacheDir, "cfdevd")
		timeSyncSocket = filepath.Join(d.Config.StateLinuxkit, "00000003.0000f3a4")
		cmd            = exec.Command("sudo", "-S", executablePath, "install", "--timesyncSock", timeSyncSocket)
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
