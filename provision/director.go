package provision

import (
	"bytes"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/runner"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	vpnkitInternalIP   = "192.168.65.3"
	vpnkitNameserverIP = "192.168.65.1"
	kvmNameserverIP    = "192.168.122.1"
)

func (c *Controller) DeployBosh() error {
	var (
		// For now we determine if we have a BOSH Director with credhub deployed
		// by looking to see if a creds.yml is present or not
		// This is definitely not the most expressive solution
		// and should be improved..
		credsPath        = filepath.Join(c.Config.StateBosh, "creds.yml")
		directorPath     = filepath.Join(c.Config.StateBosh, "director.yml")
		cloudConfigPath  = filepath.Join(c.Config.StateBosh, "cloud-config.yml")
		stateJSONPath    = filepath.Join(c.Config.StateBosh, "state.json")
		boshRunner       = runner.NewBosh(c.Config)
		crehubIsDeployed = doesNotExist(credsPath)
	)

	ip, err := driver.IP(c.Config)
	if err != nil {
		return err
	}

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-bosh.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	key, err := ioutil.ReadFile(filepath.Join(c.Config.StateDir, "id_rsa"))
	if err != nil {
		return err
	}

	s, err := NewSSH(ip, "9992", key, 20*time.Second, logFile, logFile)
	if err != nil {
		return err
	}
	defer s.Close()

	if runtime.GOOS == "linux" {
		directorContents, err := ioutil.ReadFile(directorPath)
		if err != nil {
			return err
		}

		directorContents = bytes.Replace(directorContents, []byte(vpnkitInternalIP+":9999"), []byte(ip+":9999"), -1)

		directorContents = bytes.Replace(directorContents, []byte(vpnkitNameserverIP), []byte(kvmNameserverIP), -1)

		s.SendData(directorContents, "/bosh/director.yml")
	}

	s.SendFile(stateJSONPath, "/bosh/state.json")

	command := "/usr/local/bin/bosh --tty create-env /bosh/director.yml --state /bosh/state.json"

	if !crehubIsDeployed {
		s.SendFile(credsPath, "/bosh/creds.yml")

		command = command + " --vars-store /bosh/creds.yml"
	}

	// Added the time because we were seeing some delay
	// between the time the container was started
	// and the time it could access the internet
	// Find a better solution
	time.Sleep(7 * time.Second)

	s.Run(command)

	s.RetrieveFile(stateJSONPath, "/bosh/state.json")
	if s.Error != nil {
		return s.Error
	}

	if runtime.GOOS == "linux" {
		cloudConfigContents, err := ioutil.ReadFile(cloudConfigPath)
		if err != nil {
			return err
		}

		cloudConfigContents = bytes.Replace(cloudConfigContents, []byte(vpnkitNameserverIP), []byte(kvmNameserverIP), -1)

		err = ioutil.WriteFile(cloudConfigPath, cloudConfigContents, 0600)
		if err != nil {
			return err
		}

		boshRunner.Output("-n", "update-cloud-config", cloudConfigPath)
	}

	return nil
}

func doesNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
