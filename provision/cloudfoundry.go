package provision

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Controller) DeployCloudFoundry(dockerRegistries []string) error {
	cmd := exec.Command(
		"bosh", "-n",
		"-d", "cf",
		"deploy",
		filepath.Join(c.Config.CacheDir, "cf.yml"),
		"--vars-store", filepath.Join(c.Config.StateBosh, "creds.yml"))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, c.boshEnvs()...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) boshEnvs() []string {
	content, _ := ioutil.ReadFile(filepath.Join(c.Config.StateBosh, "secret"))
	secret := strings.TrimSpace(string(content))

	return []string{
		"BOSH_ENVIRONMENT=10.0.0.4",
		"BOSH_CLIENT=admin",
		"BOSH_CLIENT_SECRET=" + secret,
		"BOSH_CA_CERT=" + filepath.Join(c.Config.StateBosh, "ca.crt"),
		"BOSH_GW_HOST=10.0.0.4",
		"BOSH_GW_USER=jumpbox",
		"BOSH_GW_PRIVATE_KEY=" + filepath.Join(c.Config.StateBosh, "jumpbox.key"),
	}
}
