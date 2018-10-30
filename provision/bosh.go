package provision

import (
	"os/exec"
	"path/filepath"
)

func (c *Controller) DeployBosh() error {
	cmd := exec.Command(
		"bosh",
		"create-env",
		filepath.Join(c.Config.CacheDir,"director.yml"),
		"--state",
		filepath.Join(c.Config.StateBosh,"state.json"),
		"--vars-store",
		filepath.Join(c.Config.StateBosh,"creds.yml"))

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
