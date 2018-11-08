package provision

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func (c *Controller) DeployCloudFoundry(ui UI, dockerRegistries []string) error {
	//TODO change to call service/cf.yml
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

	config, err := c.FetchBOSHConfig()
	if err != nil {
		return err
	}

	b, err := bosh.New(config)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- cmd.Run()
	}()

	return c.report(time.Now(), ui, b, Service{
		Name: "cf",
		Deployment: "cf",
		IsErrand: false,
	}, errChan)
}
