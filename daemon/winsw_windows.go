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
	"time"
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
	LogPath     string   `xml:"logpath"`
	LogMode     string   `xml:"logmode"`
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
	err = runCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (w *WinSW) RemoveDaemon(label string) error {
	if isInstalled(label) {
		_, executablePath := getServicePaths(label, w.ServicesDir)

		cmd := exec.Command(executablePath, "uninstall")
		err := runCommand(cmd)
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
	err := runCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (w *WinSW) Stop(label string) error {
	fmt.Printf("DEBUG: ATTEMPTING TO STOP %v\n", label)

	var executablePath string
	running, _ := w.IsRunning(label)
	for running {
		fmt.Printf("DEBUG: %v IS RUNNING\n", label)
		_, executablePath = getServicePaths(label, w.ServicesDir)

		cmd := exec.Command(executablePath, "stop")
		err := runCommand(cmd)
		if err != nil {
			return err
		}
		fmt.Printf("DEBUG: %v SHOULD HAVE STOPPED\n", executablePath)
		time.Sleep(2 * time.Second)
		running, _ = w.IsRunning(label)
	}
	return nil
}

func (w *WinSW) IsRunning(label string) (bool, error) {
	if !isInstalled(label) {
		fmt.Printf("DEBUG: %v IS NOT INSTALLED\n", label)
		return false, nil
	}

	_, executablePath := getServicePaths(label, w.ServicesDir)
	cmd := exec.Command(executablePath, "status")

	output, err := cmd.Output()

	fmt.Printf("DEBUG: STATUS: %v\n", string(output))

	if err != nil {
		return false, err
	}

	isRunning := strings.Contains(string(output), "Started") || strings.Contains(string(output), "Running")
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
		LogPath:     filepath.Dir(spec.StdoutPath),
		LogMode:     "rotate",
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

func runCommand(command *exec.Cmd) error {
	output, err := command.CombinedOutput()
	fmt.Printf("DEBUG: OUTPUT FROM TRYING TO STOP SERVICE: %v\n", output)
	if err != nil {
		return fmt.Errorf("Failed to execute %s, %v: %s: %s", command.Path, command.Args, err, string(output))
	}

	return nil
}
