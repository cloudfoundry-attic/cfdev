package bosh

import (
	"code.cloudfoundry.org/cfdev/config"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	Preparing     = "preparing"
	Deploying     = "deploying"
	RunningErrand = "running-errand"
)

type Bosh struct {
	envs []string
	cfg  config.Config
}

type VMProgress struct {
	State    string
	Total    int
	Done     int
	Duration time.Duration
}

func New(cfg config.Config) *Bosh {
	return &Bosh{
		envs: Envs(cfg),
		cfg:  cfg,
	}
}

func (b *Bosh) GetVMProgress(start time.Time, deploymentName string, isErrand bool) VMProgress {
	if isErrand {
		return VMProgress{State: RunningErrand, Duration: time.Now().Sub(start)}
	}

	executable := filepath.Join(b.cfg.BinaryDir, "bosh")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}

	command := exec.Command(executable, "--tty", "-d", deploymentName, "instances", "--json")
	command.Env = append(os.Environ(), b.envs...)
	output, err := command.Output()
	if err != nil {
		return VMProgress{State: Preparing, Duration: time.Now().Sub(start)}
	}

	var result struct {
		Tables []struct {
			Rows []struct {
				State string `json:"process_state"`
			}
		}
	}

	err = json.Unmarshal(output, &result)
	if err != nil || len(result.Tables) == 0 {
		return VMProgress{State: Preparing, Duration: time.Now().Sub(start)}
	}

	var (
		instances = result.Tables[0].Rows
		total     = len(instances)
		numDone   = 0
	)

	for _, i := range instances {
		if i.State == "running" {
			numDone++
		}
	}

	return VMProgress{State: Deploying, Total: total, Done: numDone, Duration: time.Now().Sub(start)}
}
