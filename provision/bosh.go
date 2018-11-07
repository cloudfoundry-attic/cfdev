package provision

import (
	"os"
	"os/exec"
	"path/filepath"
)

func (c *Controller) DeployBosh() error {
	cmd := exec.Command(
		filepath.Join(c.Config.CacheDir, "bosh"),
		"--tty", "create-env",
		filepath.Join(c.Config.CacheDir, "director.yml"),
		"--state",
		filepath.Join(c.Config.StateBosh, "state.json"),
		"--vars-store",
		filepath.Join(c.Config.StateBosh, "creds.yml"))

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-bosh.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
