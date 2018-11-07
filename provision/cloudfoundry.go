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
		"bosh", "--tty", "-n",
		"-d", "cf",
		"deploy",
		filepath.Join(c.Config.CacheDir, "cf.yml"),
		"--vars-store", filepath.Join(c.Config.StateBosh, "creds.yml"))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, c.boshEnvs()...)

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-cf.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd.Run()
}

func (c *Controller) boshEnvs() []string {
	content, _ := ioutil.ReadFile(filepath.Join(c.Config.StateBosh, "secret"))
	secret := strings.TrimSpace(string(content))

	return []string{
		"BOSH_ENVIRONMENT=10.144.0.4",
		"BOSH_CLIENT=admin",
		"BOSH_CLIENT_SECRET=" + secret,
		"BOSH_CA_CERT=" + filepath.Join(c.Config.StateBosh, "ca.crt"),
		"BOSH_GW_HOST=10.144.0.4",
		"BOSH_GW_USER=jumpbox",
		"BOSH_GW_PRIVATE_KEY=" + filepath.Join(c.Config.StateBosh, "jumpbox.key"),
		"SERVICES_DIR=" + c.Config.ServicesDir,
		"CACHE_DIR=" + c.Config.CacheDir,
		"BOSH_STATE=" + c.Config.StateBosh,
	}
}
