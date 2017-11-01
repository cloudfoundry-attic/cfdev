package user

import (
	"os"
	"os/user"
	"path/filepath"
)

func CFDevHome() (string, error) {
	path := os.Getenv("CFDEV_HOME")

	if path != "" {
		return path, os.MkdirAll(path, 0755)
	}

	u, err := user.Current()

	if err != nil {
		return "", err
	}

	path = filepath.Join(u.HomeDir, ".cfdev")
	return path, os.MkdirAll(path, 0755)
}
