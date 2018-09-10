package runner

import (
	"fmt"
	"os/exec"
)

type Powershell struct {}

func (p *Powershell) Output(command string) (string, error) {
	output, err := exec.Command("powershell.exe", "-Command", command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute: powershell.exe -Command %q: %s: %s", command, err, output)
	}

	return string(output), nil
}
