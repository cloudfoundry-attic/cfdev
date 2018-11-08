package shell

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"

	"bytes"

	"runtime"
)

type Environment struct {
}

func (e *Environment) Prepare(config bosh.Config) (string, error) {
	order := []string{
		"BOSH_ENVIRONMENT",
		"BOSH_CLIENT",
		"BOSH_CLIENT_SECRET",
		"BOSH_CA_CERT",
		"BOSH_GW_HOST",
		"BOSH_GW_PRIVATE_KEY",
		"BOSH_GW_USER",
	}

	values := map[string]string{
		"BOSH_ENVIRONMENT":    config.DirectorAddress,
		"BOSH_CLIENT":         config.AdminUsername,
		"BOSH_CLIENT_SECRET":  config.AdminPassword,
		"BOSH_CA_CERT":        config.CACertificate,
		"BOSH_GW_HOST":        config.GatewayHost,
		"BOSH_GW_PRIVATE_KEY": config.GatewayPrivateKey,
		"BOSH_GW_USER":        config.GatewayUsername,
	}

	var output bytes.Buffer

	for _, envvar := range os.Environ() {
		if strings.HasPrefix(envvar, "BOSH_") {
			envvar = strings.Split(envvar, "=")[0]
			if runtime.GOOS != "windows" {
				fmt.Fprintf(&output, "unset %s;\n", envvar)
			} else {
				fmt.Fprintf(&output, "Remove-Item Env:%s;\n", envvar)
			}
		}
	}

	for _, name := range order {
		if runtime.GOOS != "windows" {
			fmt.Fprintf(&output, "export %s=\"%s\";\n", name, values[name])
		} else {
			fmt.Fprintf(&output, "$env:%s=\"%s\";\n", name, values[name])
		}
	}

	return strings.TrimSpace(output.String()), nil
}
