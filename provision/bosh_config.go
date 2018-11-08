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

	return bosh.Config{
		AdminUsername:     "admin",
		AdminPassword:     secret,
		CACertificate:     filepath.Join(c.Config.StateBosh, "ca.crt"),
		DirectorAddress:   "10.144.0.4",
		GatewayHost:       "10.144.0.4",
		GatewayPrivateKey: filepath.Join(c.Config.StateBosh, "jumpbox.key"),
		GatewayUsername:   "jumpbox",
	}, nil
}