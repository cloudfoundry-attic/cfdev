package runner

import "os/exec"

type Runner struct {}

func (r *Runner) Output(command string, arg ...string) ([]byte, error) {
	return exec.Command(command, arg...).Output()
}
