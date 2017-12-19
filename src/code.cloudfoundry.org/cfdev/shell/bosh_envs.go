package shell

import (
	"fmt"

	"bytes"

	"code.cloudfoundry.org/cfdev/garden"
)

func FormatConfig(config garden.BOSHConfiguration) (string, error) {
	order := []string{
		"BOSH_ENVIRONMENT",
		"BOSH_CLIENT",
		"BOSH_CLIENT_SECRET",
		"BOSH_CA_CERT",
		"BOSH_GW_PRIVATE_KEY",
	}

	values := map[string]string{
		"BOSH_ENVIRONMENT":    config.IPAddress,
		"BOSH_CLIENT":         config.AdminUsername,
		"BOSH_CLIENT_SECRET":  config.AdminPassword,
		"BOSH_CA_CERT":        config.CACertificate,
		"BOSH_GW_PRIVATE_KEY": config.SSHPrivateKey,
	}

	var output bytes.Buffer
	for _, name := range order {
		fmt.Fprintf(&output, "export %v=\"%v\"\n", name, values[name])
	}
	return output.String(), nil
}
