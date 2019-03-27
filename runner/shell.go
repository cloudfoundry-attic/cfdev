package runner

import "os/exec"

type Shell struct {}

func (s *Shell) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}