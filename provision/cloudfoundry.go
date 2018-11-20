package provision

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (c *Controller) DeployCloudFoundry(ui UI, dockerRegistries []string) error {
	script := "deploy-cf"
	if runtime.GOOS == "windows" {
		script = "deploy-cf.ps1"
	}

	cmd := exec.Command(filepath.Join(c.Config.ServicesDir, script))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, c.boshEnvs()...)

	var arr []string
	for _, registry := range dockerRegistries {
		arr = append(arr, fmt.Sprintf(`%q`, registry))
	}

	cmd.Env = append(cmd.Env, `DOCKER_REGISTRIES=[`+strings.Join(arr, ",")+"]")

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
		Name:       "cf",
		Deployment: "cf",
		IsErrand:   false,
	}, errChan)
}
