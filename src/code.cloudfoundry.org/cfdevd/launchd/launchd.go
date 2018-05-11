package launchd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"code.cloudfoundry.org/cfdevd/launchd/models"
)

type Launchd struct {
	PListDir string
}

func New(pListDir string) *Launchd {
	if pListDir == "" {
		pListDir = "/Library/LaunchDaemons"
	}
	return &Launchd{
		PListDir: pListDir,
	}
}

func (l *Launchd) AddDaemon(spec models.DaemonSpec) error {
	plistPath := filepath.Join(l.PListDir, spec.Label+".plist")
	l.remove(spec.Label)
	if err := l.writePlist(spec, plistPath); err != nil {
		return err
	}
	return l.load(plistPath)
}

func (l *Launchd) RemoveDaemon(label string) error {
	plistPath := filepath.Join(l.PListDir, label+".plist")
	loaded, err := l.isLoaded(label)
	if err != nil {
		return err
	}
	if loaded {
		if err := l.remove(label); err != nil {
			return err
		}
	}

	err = os.Remove(plistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

func (l *Launchd) Start(label string) error {
	cmd := exec.Command("launchctl", "start", label)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) Stop(label string) error {
	if running, _ := l.IsRunning(label); !running {
		return nil
	}
	cmd := exec.Command("launchctl", "stop", label)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) list() (string, error) {
	out, err := exec.Command("launchctl", "list").Output()
	return string(out), err
}

func (l *Launchd) IsRunning(label string) (bool, error) {
	out, err := l.list()
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		cols := strings.Fields(line)
		if len(cols) >= 3 && cols[2] == label {
			return cols[0] != "-", nil
		}
	}
	return false, nil
}

func (l *Launchd) isLoaded(label string) (bool, error) {
	out, err := l.list()
	if err != nil {
		return false, err
	}
	return strings.Contains(out, label), nil
}

func (l *Launchd) load(plistPath string) error {
	cmd := exec.Command("launchctl", "load", "-F", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) remove(label string) error {
	cmd := exec.Command("launchctl", "remove", label)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) writePlist(spec models.DaemonSpec, dest string) error {
	tmplt := template.Must(template.New("plist").Parse(plistTemplate))
	plist, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer plist.Close()
	return tmplt.Execute(plist, spec)
}

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>Program</key>
  <string>{{.Program}}</string>
  <key>ProgramArguments</key>
  <array>
  {{range .ProgramArguments}}
    <string>{{.}}</string>
	{{end}}
  </array>
  <key>RunAtLoad</key>
  <{{.RunAtLoad}}/>
  {{if .Sockets}}
  <key>Sockets</key>
  <dict>
	  {{range $name, $path := .Sockets}}
    <key>{{$name}}</key>
    <dict>
      <key>SockPathMode</key>
      <integer>438</integer>
      <key>SockPathName</key>
      <string>{{$path}}</string>
    </dict>
		{{end}}
  </dict>
  {{end}}
  {{if .StdoutPath}}
	<key>StandardOutPath</key>
	<string>{{.StdoutPath}}</string>
  {{end}}
  {{if .StderrPath}}
	<key>StandardErrorPath</key>
	<string>{{.StderrPath}}</string>
  {{end}}
</dict>
</plist>
`
