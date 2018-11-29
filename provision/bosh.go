package provision

import (
	"code.cloudfoundry.org/cfdev/ssh"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func (c *Controller) DeployBosh() error {
	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-bosh.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	key, err := ioutil.ReadFile(filepath.Join(c.Config.CacheDir, "id_rsa"))
	if err != nil {
		return err
	}
	s := ssh.SSH{}

	srcDst := []string{
		filepath.Join(c.Config.CacheDir, "director.yml"),
		filepath.Join(c.Config.StateBosh, "state.json"),
		filepath.Join(c.Config.StateBosh, "creds.yml"),
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

	command := fmt.Sprintf("%s --tty create-env %s --state %s --vars-store %s",
		"/bosh/bosh",
		"director.yml",
		"state.json",
		"creds.yml")

	// TODO: Added the time because we were seeing some delay between the time the container
	// was started and the time it could access the internet
	// Find a better solution
	time.Sleep(7 * time.Second)

	err = s.RunSSHCommand(
		command,
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
