package launchd

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type program struct {
	executable string
	args       []string
}

type Config struct {
	XMLName     xml.Name `xml:"configuration"`
	Id          string   `xml:"id"`
	Name        string   `xml:"name"`
	Description string   `xml:"description"`
	Executable  string   `xml:"executable"`
	Arguments   string   `xml:"arguments"`
	StartMode   string   `xml:"startmode"`
}

func (l *Launchd) AddDaemon(spec DaemonSpec) error {


	serviceDst, executablePath, err := copyBinary(spec)
	if err != nil {
		return err
	}

	err = createXml(serviceDst, spec)
	if err != nil {
		return err
	}

	cmd := exec.Command(executablePath, "install")
	err = cmd.Start()
	if err != nil {
		return err
	}


	err = cmd.Wait()
	if err != nil {
		return err
	}


	return nil
}

func (l *Launchd) RemoveDaemon(spec DaemonSpec) error {
	if isInstalled(spec) {
		_, executablePath := getServicePaths(spec.Label, spec.CfDevHome)

		cmd := exec.Command(executablePath, "uninstall")
		err := cmd.Start()
		if err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (l *Launchd) Start(spec DaemonSpec) error {
	_, executablePath := getServicePaths(spec.Label, spec.CfDevHome)


	cmd := exec.Command(executablePath, "start")
	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (l *Launchd) Stop(spec DaemonSpec) error {
	if running, _ := l.IsRunning(spec); running {

		_, executablePath := getServicePaths(spec.Label, spec.CfDevHome)

		cmd := exec.Command(executablePath, "stop")
		err := cmd.Start()
		if err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (l *Launchd) IsRunning(spec DaemonSpec) (bool, error) {
	_, executablePath := getServicePaths(spec.Label, spec.CfDevHome)
	cmd := exec.Command(executablePath, "status")

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	isRunning := strings.Contains(string(output), "Started")
	return isRunning, nil
}

func copyBinary(spec DaemonSpec) (string, string, error) {
	serviceDst, executablePath := getServicePaths(spec.Label, spec.CfDevHome)

	err := os.MkdirAll(serviceDst, 0666)
	if err != nil {
		return "", "", err
	}

	winswPath := filepath.Join(spec.CfDevHome, "cache", "winsw.exe")
	winswData, err := ioutil.ReadFile(winswPath)
	if err != nil {
		return "", "", err
	}

	err = ioutil.WriteFile(executablePath, winswData, 0666)
	if err != nil {
		return "", "", err
	}

	return serviceDst, executablePath, nil
}

func createXml(serviceDst string, spec DaemonSpec) error {
	file, err := os.Create(filepath.Join(serviceDst, spec.Label+".xml"))
	if err != nil {
		return err
	}

	config := &Config{
		Id:          spec.Label,
		Name:        spec.Label,
		Description: spec.Label,
		Executable:  spec.Program,
		Arguments:   strings.Join(spec.ProgramArguments[:], " "),
		StartMode:   "Manual",
	}
	configWriter := io.Writer(file)

	enc := xml.NewEncoder(configWriter)
	enc.Encode(config)
	file.Close()

	return nil
}

func getServicePaths(label string, cfDevHome string) (string, string) {

	serviceDst := filepath.Join(cfDevHome, "winservice", label)
	executablePath := filepath.Join(serviceDst, label+".exe")

	return serviceDst, executablePath
}

func isInstalled(spec DaemonSpec) bool{
	command := exec.Command("powershell.exe", "-C", fmt.Sprintf(`Get-Service | Where-Object {$_.Name -eq "%s"}`, spec.Label))
	temp, _ := command.Output()
	output := strings.TrimSpace(string(temp))

	return output != ""

}