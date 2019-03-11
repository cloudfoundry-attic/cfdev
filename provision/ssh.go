package provision

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSH struct {}

type SSHAddress struct {
	IP   string
	Port string
}

func (s *SSH) CopyFile(filePath string, remoteFilePath string, address SSHAddress, privateKey []byte, timeout time.Duration, stdout io.Writer, stderr io.Writer) error {
	client, session, err := s.newSession(address, privateKey, timeout)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	l, err := os.Open(filePath)
	defer l.Close()

	command := fmt.Sprintf("/usr/bin/scp -qt %s", filepath.Dir(remoteFilePath))
	contentsBytes, _ := ioutil.ReadAll(l)
	bytesReader := bytes.NewReader(contentsBytes)

	go func() {
		w, _ := session.StdinPipe()

		fmt.Fprintln(w, "C0755", int64(len(contentsBytes)), remoteFilePath)
		_, err = io.Copy(w, bytesReader)
		if err != nil {
			fmt.Print(err)
		}

		fmt.Fprintln(w, "\x00")

		defer w.Close()
	}()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(command)
}

func (s *SSH) RetrieveFile(filePath string, remoteFilePath string, address SSHAddress, privateKey []byte, timeout time.Duration) error {
	client, session, err := s.newSession(address, privateKey, timeout)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	session.Stdout = f
	return session.Run("cat " + remoteFilePath)
}

func (s *SSH) RunSSHCommand(command string, addresses SSHAddress, privateKey []byte, timeout time.Duration, stdout io.Writer, stderr io.Writer) (err error) {
	client, session, err := s.newSession(addresses, privateKey, timeout)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(command)
}

func (s *SSH) WaitForSSH(addresses SSHAddress, privateKey []byte, timeout time.Duration) error {
	client, err := s.waitForSSH(addresses, privateKey, timeout)
	if err == nil {
		client.Close()
	}
	return err
}

func (s *SSH) newSession(addresses SSHAddress, privateKey []byte, timeout time.Duration) (*ssh.Client, *ssh.Session, error) {
	client, err := s.waitForSSH(addresses, privateKey, timeout)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

func (*SSH) waitForSSH(address SSHAddress, privateKey []byte, timeout time.Duration) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not parse private key: %s", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}

	clientChan := make(chan *ssh.Client, 1)
	errorChan := make(chan error, 1)
	doneChan := make(chan bool)

	go func(ip string, port string) {
		var client *ssh.Client
		var dialErr error
		timeoutChan := time.After(timeout)
		for {
			select {
			case <-timeoutChan:
				clientChan <- nil
				errorChan <- fmt.Errorf("ssh connection timed out: %s", dialErr)
				return
			case <-doneChan:
				return
			default:
				if client, dialErr = ssh.Dial("tcp", ip+":"+port, config); dialErr == nil {
					clientChan <- client
					errorChan <- nil
					return
				}
				time.Sleep(time.Second)
			}
		}
	}(address.IP, address.Port)

	client := <-clientChan
	err = <-errorChan
	close(doneChan)
	return client, err
}
