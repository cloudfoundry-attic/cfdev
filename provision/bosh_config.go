package provision

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"
)

func (c *Controller) FetchBOSHConfig() (bosh.Config, error) {
	content, err := ioutil.ReadFile(filepath.Join(c.Config.StateBosh, "secret"))
	if err != nil {
		return bosh.Config{}, err
	}

	secret := strings.TrimSpace(string(content))

	content, err = ioutil.ReadFile(filepath.Join(c.Config.StateBosh, "ca.crt"))
	if err != nil {
		return bosh.Config{}, err
	}

	caCert := strings.TrimSpace(string(content))

	return bosh.Config{
		AdminUsername:     "admin",
		AdminPassword:     secret,
		CACertificate:     caCert,
		DirectorAddress:   "10.144.0.4",
		GatewayHost:       "10.144.0.4",
		GatewayPrivateKey: filepath.Join(c.Config.StateBosh, "jumpbox.key"),
		GatewayUsername:   "jumpbox",
	}, nil
}

func (c *Controller) boshEnvs() []string {
	config, _ := c.FetchBOSHConfig()

	return []string{
		"BOSH_ENVIRONMENT=" + config.DirectorAddress,
		"BOSH_CLIENT=" + config.AdminUsername,
		"BOSH_CLIENT_SECRET=" + config.AdminPassword,
		"BOSH_CA_CERT=" + config.CACertificate,
		"BOSH_GW_HOST=" + config.GatewayHost,
		"BOSH_GW_USER=" + config.GatewayUsername,
		"BOSH_GW_PRIVATE_KEY=" + config.GatewayPrivateKey,
		"SERVICES_DIR=" + c.Config.ServicesDir,
		"CACHE_DIR=" + c.Config.CacheDir,
		"BOSH_STATE=" + c.Config.StateBosh,
	}
}
