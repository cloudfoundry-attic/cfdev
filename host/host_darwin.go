package host

import (
	"fmt"
	"os/exec"
	"strings"
)

func (*Host) CheckRequirements() error {
	return nil
}
func (h *Host) Version() (string, error) {
	name, err := exec.Command("sw_vers", "-productName").Output()
	if err != nil {
		return "", err
	}

	version, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", strings.TrimSpace(string(name)), strings.TrimSpace(string(version))), nil
}
