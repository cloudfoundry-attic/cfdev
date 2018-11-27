package provision

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func (c *Controller) DeployService(service Service) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", filepath.Join(c.Config.ServicesDir, service.Script+".ps1"))
	} else {
		cmd = exec.Command(filepath.Join(c.Config.ServicesDir, service.Script))
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, bosh.Envs(c.Config)...)

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-"+strings.ToLower(service.Name)+".log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd.Run()
}

type Service struct {
	Name          string `yaml:"name"`
	Flagname      string `yaml:"flag_name"`
	DefaultDeploy bool   `yaml:"default_deploy"`
	Handle        string `yaml:"handle"` //TODO <-- remove
	Script        string `yaml:"script"`
	Deployment    string `yaml:"deployment"`
	IsErrand      bool   `yaml:"errand"`
}
