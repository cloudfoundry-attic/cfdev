package provision

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

//go:generate mockgen -package mocks -destination mocks/runner.go code.cloudfoundry.org/cfdev/provision Runner
type Runner interface {
	Output(cmd *exec.Cmd) ([]byte, error)
}

type Instance struct {
	ID           string `json:"instance"`
	Process      string `json:"process"`
	ProcessState string `json:"process_state"`
}

type Bosh struct {
	envs   []string
	cfg    config.Config
	Runner Runner
}

type VMProgress struct {
	State    string
	Total    int
	Done     int
	Duration time.Duration
}

func NewBosh(cfg config.Config, runner Runner, envs []string) *Bosh {
	return &Bosh{
		cfg:  cfg,
		envs: envs,
		Runner: runner,
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

	command := exec.Command(executable, "--tty", "-d", deploymentName, "instances", "--ps", "--json")
	command.Env = append(os.Environ(), b.envs...)
	output, err := b.Runner.Output(command)
	if err != nil {
		return VMProgress{State: Preparing, Duration: time.Now().Sub(start)}
	}

	var result struct {
		Tables []struct {
			Instances []Instance `json:"Rows"`
		} `json:"Tables"`
	}

	err = json.Unmarshal(output, &result)
	if err != nil || len(result.Tables) == 0 || len(result.Tables[0].Instances) == 0 {
		return VMProgress{State: Preparing, Duration: time.Now().Sub(start)}
	}

	numDone, total := parseResults(result.Tables[0].Instances)
	return VMProgress{State: Deploying, Total: total, Done: numDone, Duration: time.Now().Sub(start)}
}

func parseResults(instances []Instance) (int, int) {
	var (
		uniqInstances    = map[string]bool{}
		isCompletedCount = func() int {
			var count int
			for _, v := range uniqInstances {
				if v {
					count++
				}
			}
			return count
		}
		allProcessesRunning = func(id string) bool {
			var hasAtLeastOneProcess bool
			for _, i := range instances {
				if i.ID == id && i.Process != "" {
					hasAtLeastOneProcess = true

					if i.ProcessState != "running" {
						return false
					}
				}
			}
			return hasAtLeastOneProcess
		}
	)

	for _, i := range instances {
		uniqInstances[i.ID] = false
	}

	for k, _ := range uniqInstances {
		if allProcessesRunning(k) {
			uniqInstances[k] = true
		}
	}

	return isCompletedCount(), len(uniqInstances)
}
