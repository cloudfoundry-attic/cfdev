package daemon

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

type WinSW struct {
	BinaryPath  string
	ServicesDir string
}

func NewWinSW(cfDevHome string) *WinSW {
	return &WinSW{
		BinaryPath:  filepath.Join(cfDevHome, "cache", "winsw.exe"),
		ServicesDir: filepath.Join(cfDevHome, "winservice"),
	}
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

func RunCommand(command *exec.Cmd) error {
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to execute %s, %v: %s: %s", command.Path, command.Args, err, string(output))
	}

	return nil
}

func (w *WinSW) AddDaemon(spec DaemonSpec) error {
	serviceDst, executablePath := getServicePaths(spec.Label, w.ServicesDir)
	err := os.MkdirAll(serviceDst, 0666)
	if err != nil {
		return err
	}

	err = copyBinary(w.BinaryPath, executablePath)
	if err != nil {
		return err
	}

	err = createXml(serviceDst, spec)
	if err != nil {
		return err
	}

	cmd := exec.Command(executablePath, "install")
	err = RunCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (w *WinSW) RemoveDaemon(label string) error {
	if isInstalled(label) {
		_, executablePath := getServicePaths(label, w.ServicesDir)

		cmd := exec.Command(executablePath, "uninstall")
		err := RunCommand(cmd)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (w *WinSW) Start(label string) error {
	_, executablePath := getServicePaths(label, w.ServicesDir)

	cmd := exec.Command(executablePath, "start")
	err := RunCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (w *WinSW) Stop(label string) error {
	if running, _ := w.IsRunning(label); running {

		_, executablePath := getServicePaths(label, w.ServicesDir)

		cmd := exec.Command(executablePath, "stop")
		err := RunCommand(cmd)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (w *WinSW) IsRunning(label string) (bool, error) {
	_, executablePath := getServicePaths(label, w.ServicesDir)
	cmd := exec.Command(executablePath, "status")

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	isRunning := strings.Contains(string(output), "Started")
	return isRunning, nil
}

func copyBinary(src, dst string) error {
	winswData, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, winswData, 0666)
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

func getServicePaths(label string, servicesDir string) (string, string) {
	serviceDst := filepath.Join(servicesDir, label)
	executablePath := filepath.Join(serviceDst, label+".exe")

	return serviceDst, executablePath
}

func isInstalled(label string) bool {
	command := exec.Command("powershell.exe", "-C", fmt.Sprintf(`Get-Service | Where-Object {$_.Name -eq "%s"}`, label))
	temp, _ := command.Output()
	output := strings.TrimSpace(string(temp))

	return output != ""
}
