package provision

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (c *Controller) DeployService(service Service) error {
	cmd := exec.Command(filepath.Join(c.Config.ServicesDir, service.Script))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, c.boshEnvs()...)

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
