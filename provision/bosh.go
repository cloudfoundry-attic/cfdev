package provision

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (c *Controller) DeployBosh() error {
	var (
		// For now we determine if we have a BOSH Director with credhub deployed
		// by looking to see if a creds.yml is present or not
		// This is definitely not the most expressive solution
		// and should be improved..
		s                = SSH{}
		credsPath        = filepath.Join(c.Config.StateBosh, "creds.yml")
		directorPath     = filepath.Join(c.Config.StateBosh, "director.yml")
		crehubIsDeployed = doesNotExist(credsPath)
	)

	ip, err := c.fetchIP()
	if err != nil {
		return err
	}

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-bosh.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	key, err := ioutil.ReadFile(filepath.Join(c.Config.StateDir, "id_rsa"))
	if err != nil {
		return err
	}

	srcDst := []string{filepath.Join(c.Config.StateBosh, "state.json")}

	// If we are on a linux platform
	// We need to replace the eth0 address
	// because it will be dynamic
	if runtime.GOOS == "linux" {
		err = s.RetrieveFile(
			directorPath,
			"/bosh/director.yml",
			SSHAddress{IP: ip, Port: "9992"},
			key,
			20*time.Second)

		contents, err := ioutil.ReadFile(directorPath)
		if err != nil {
			return err
		}

		contents = bytes.Replace(contents, []byte("192.168.65.3:9999"), []byte(ip+":9999"), -1)

		err = ioutil.WriteFile(directorPath, contents, 0600)
		if err != nil {
			return err
		}

		s.CopyFile(directorPath, "/bosh/director.yml", SSHAddress{
			IP:   ip,
			Port: "9992",
		},
			key,
			20*time.Second,
			logFile,
			logFile)
	}

	if !crehubIsDeployed {
		srcDst = append(srcDst, credsPath)
	}

	for _, item := range srcDst {
		s.CopyFile(item, filepath.Base(item), SSHAddress{
			IP:   ip,
			Port: "9992",
		},
			key,
			20*time.Second,
			logFile,
			logFile)
	}

	command := []string{
		"/usr/local/bin/bosh", "--tty",
		"create-env", "/bosh/director.yml", "--state", "state.json",
	}

	if !crehubIsDeployed {
		command = append(command, "--vars-store", "creds.yml")
	}

	// Added the time because we were seeing some delay
	// between the time the container was started
	// and the time it could access the internet
	// Find a better solution
	time.Sleep(7 * time.Second)

	err = s.RunSSHCommand(
		strings.Join(command, " "),
		SSHAddress{
			IP:   ip,
			Port: "9992",
		},
		key,
		20*time.Second,
		logFile,
		logFile,
	)

	if err != nil {
		return err
	}

	return s.RetrieveFile(
		filepath.Join(c.Config.StateBosh, "state.json"),
		"/root/state.json",
		SSHAddress{IP: ip, Port: "9992"},
		key,
		20*time.Second)
}

func doesNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
