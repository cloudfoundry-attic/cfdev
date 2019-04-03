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
	vpnkitNameserverIP = "192.168.65.1"
	vpnkitHostIP       = "192.168.65.2"
	vpnkitInternalIP   = "192.168.65.3"
	kvmNameserverIP    = "192.168.122.1"
)

func (c *Controller) DeployBosh() error {
	var (
		credsPath           = filepath.Join(c.Config.StateBosh, "creds.yml")
		directorPath        = filepath.Join(c.Config.StateBosh, "director.yml")
		cloudConfigPath     = filepath.Join(c.Config.StateBosh, "cloud-config.yml")
		dnsConfigPath       = filepath.Join(c.Config.StateBosh, "dns.yml")
		opsManDnsConfigPath = filepath.Join(c.Config.StateBosh, "ops-manager-dns-runtime.yml")
		stateJSONPath       = filepath.Join(c.Config.StateBosh, "state.json")
		boshRunner          = runner.NewBosh(c.Config)
		credhubIsDeployed   = func() bool {
			// For now we determine if we have a BOSH Director with credhub deployed
			// by looking to see if a creds.yml is present or not
			// This is definitely not the most expressive solution
			// and should be improved..
			_, err := os.Stat(credsPath)
			return os.IsNotExist(err)
		}
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

	directorContents, err := ioutil.ReadFile(directorPath)
	if err != nil {
		return err
	}

	if runtime.GOOS == "linux" {
		directorContents = bytes.Replace(directorContents, []byte(vpnkitInternalIP+":9999"), []byte(ip+":9999"), -1)

		directorContents = bytes.Replace(directorContents, []byte(vpnkitNameserverIP), []byte(kvmNameserverIP), -1)
	}

	s.SendData(directorContents, "director.yml")

	s.SendFile(stateJSONPath, "state.json")

	command := "/usr/local/bin/bosh --tty create-env director.yml --state state.json"

	if !credhubIsDeployed() {
		s.SendFile(credsPath, "creds.yml")

		command = command + " --vars-store creds.yml"
	}

	// Added the time because we were seeing some delay
	// between the time the container was started
	// and the time it could access the internet
	// Find a better solution
	time.Sleep(7 * time.Second)

	s.Run(command)

	s.RetrieveFile(stateJSONPath, "state.json")
	if s.Error != nil {
		return s.Error
	}

	if runtime.GOOS == "linux" {
		err = c.updateCloudConfig(boshRunner, cloudConfigPath)
		if err != nil {
			return err
		}

		err = c.updateDNSRuntime(boshRunner, dnsConfigPath)
		if err != nil {
			return err
		}

		err = c.updateOpsManDNSRuntime(boshRunner, opsManDnsConfigPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) updateCloudConfig(boshRunner *runner.Bosh, path string) error {
	cloudConfigContents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	cloudConfigContents = bytes.Replace(cloudConfigContents, []byte(vpnkitNameserverIP), []byte(kvmNameserverIP), -1)

	err = ioutil.WriteFile(path, cloudConfigContents, 0600)
	if err != nil {
		return err
	}

	_, err = boshRunner.Output("-n", "update-cloud-config", path)
	return err
}

func (c *Controller) updateDNSRuntime(boshRunner *runner.Bosh, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// dns.yml does not exist
		// skipping in favor of next DNS runtime method updater
		return nil
	}

	dnsConfigContents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	dnsConfigContents = bytes.Replace(dnsConfigContents, []byte(vpnkitHostIP), []byte(kvmNameserverIP), -1)

	err = ioutil.WriteFile(path, dnsConfigContents, 0600)
	if err != nil {
		return err
	}

	_, err = boshRunner.Output("-n", "update-runtime-config", path)
	return err
}

func (c *Controller) updateOpsManDNSRuntime(boshRunner *runner.Bosh, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// dns.yml does not exist
		// skipping in favor of next DNS runtime method updater
		return nil
	}

	dnsConfigContents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	dnsConfigContents = bytes.Replace(dnsConfigContents, []byte(vpnkitHostIP), []byte(kvmNameserverIP), -1)

	err = ioutil.WriteFile(path, dnsConfigContents, 0600)
	if err != nil {
		return err
	}

	_, err = boshRunner.Output("-n", "update-config", "--name", "ops_manager_dns_runtime", "--type", "runtime", path)
	return err
}
