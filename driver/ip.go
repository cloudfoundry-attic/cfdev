// +build !linux

package driver

import "code.cloudfoundry.org/cfdev/config"

func IP(cfg config.Config) (string, error) {
		return "127.0.0.1", nil
}
