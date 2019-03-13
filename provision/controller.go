package provision

import (
	"code.cloudfoundry.org/cfdev/config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aemengo/bosh-runc-cpi/client"
	"io"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type Controller struct {
	Config config.Config
}

func NewController(config config.Config) *Controller {
	return &Controller{
		Config: config,
	}
}

func (c *Controller) Ping(duration time.Duration) error {
	var (
		ticker  = time.NewTicker(time.Second)
		timeout = time.After(duration)
		err     error
	)

	for {
		select {
		case <-ticker.C:
			var ip string
			ip, err = c.fetchIP()
			if err != nil {
				continue
			}

			err = client.Ping(context.Background(), ip+":9999")
			if err == nil {
				return nil
			}
		case <-timeout:
			return err
		}
	}
}

func (c *Controller) fetchIP() (string, error) {
	if runtime.GOOS != "linux" {
		return "127.0.0.1", nil
	}

	var (
		macAddrPath     = filepath.Join(c.Config.StateLinuxkit, "mac-addr")
		vBridgeInfoPath = filepath.Join("/var/lib/libvirt/dnsmasq/virbr0.status")
	)

	macAddr, err := ioutil.ReadFile(macAddrPath)
	if err != nil {
		return "", err
	}

	vBridgeInfo, err := ioutil.ReadFile(vBridgeInfoPath)
	if err != nil {
		return "", err
	}

	var results []struct {
		IPAddr  string `json:"ip-address"`
		MacAddr string `json:"mac-address"`
	}

	err = json.Unmarshal(vBridgeInfo, &results)
	if err != nil {
		return "", err
	}

	for _, result := range results {
		if result.MacAddr == string(macAddr) {
			return result.IPAddr, nil
		}
	}

	return "", fmt.Errorf("unable to find VM IP address from '%s'", vBridgeInfoPath)
}
