package os

import (
	"os/exec"
	"strings"
)

func (o *OS) Version() (string, error) {
	output, err := exec.Command("powershell.exe", "-Command", "[System.Environment]::OSVersion.VersionString").Output()
	return strings.TrimSpace(string(output)), err
}