package os

import (
	"code.cloudfoundry.org/cfdev/errors"
	"fmt"
	"strings"
)

func (o *OS) Version() (string, error) {

	var version string
	output, err := o.Runner.Output("sw_vers")
	if err != nil {
		return "", errors.SafeWrap(err, "failed to get OS version")
	}
	splitOutput := strings.Split(string(output), "\n")
	for _, line := range splitOutput {
		if strings.Contains(line, "ProductVersion:") {
			parseLine := strings.Split(line, "\t")
			version = parseLine[len(parseLine) -1]
			break
		}
	}

	if version == "" {
		return "", fmt.Errorf("failed to parse os version out of: %s", output)
	}

	return version, nil
}
