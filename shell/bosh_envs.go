package shell

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"

	"bytes"

	"io/ioutil"
	"path/filepath"
	"runtime"
)

type Environment struct {
	StateDir string
}

func (e *Environment) Prepare(config bosh.Config) (string, error) {
	keyPath := filepath.Join(e.StateDir, "bosh-gw-key")
	certPath := filepath.Join(e.StateDir, "bosh-ca-cert")

	err := ioutil.WriteFile(keyPath, []byte(config.GatewayPrivateKey), 0600)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(certPath, []byte(config.CACertificate), 0666)
	if err != nil {
		return "", err
	}

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
		"BOSH_CA_CERT":        certPath,
		"BOSH_GW_HOST":        config.GatewayHost,
		"BOSH_GW_PRIVATE_KEY": keyPath,
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
