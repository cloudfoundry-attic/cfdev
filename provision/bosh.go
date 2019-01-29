package provision

import (
	"code.cloudfoundry.org/cfdev/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *Controller) DeployBosh() error {
	var (
		// For now we determine if we have a BOSH Director with credhub deployed
		// by looking to see if a creds.yml is present or not
		// This is definitely not the most expressive solution
		// and should be improved..
		s                = ssh.SSH{}
		credsPath        = filepath.Join(c.Config.StateBosh, "creds.yml")
		crehubIsDeployed = doesNotExist(credsPath)
	)

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

	if !crehubIsDeployed {
		srcDst = append(srcDst, credsPath)
	}

	for _, item := range srcDst {
		s.CopyFile(item, filepath.Base(item), ssh.SSHAddress{
			IP:   "127.0.0.1",
			Port: "9992",
		},
			key,
			20*time.Second,
			logFile,
			logFile)
	}

	command := []string{
		"/bosh/bosh", "--tty",
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
		ssh.SSHAddress{
			IP:   "127.0.0.1",
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
		ssh.SSHAddress{IP: "127.0.0.1", Port: "9992"},
		key,
		20*time.Second)
}

func doesNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
