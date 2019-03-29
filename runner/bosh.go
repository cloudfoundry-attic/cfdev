package runner

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/workspace"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Bosh struct {
	config    config.Config
	workspace *workspace.Workspace
}

func NewBosh(config config.Config) *Bosh {
	return &Bosh{
		config:    config,
		workspace: workspace.New(config),
	}
}

func (b *Bosh) Output(args ...string) ([]byte, error) {
	executable := filepath.Join(b.config.BinaryDir, "bosh")

	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	command := exec.Command(executable, args...)
	command.Env = append(os.Environ(), b.workspace.Envs()...)
	return command.Output()
}
