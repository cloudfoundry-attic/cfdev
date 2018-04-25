package launchd

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type DaemonSpec struct {
	Label            string
	Program          string
	ProgramArguments []string
	RunAtLoad        bool
	Sockets          map[string]string
	StdoutPath       string
	StderrPath       string
}

type Launchd struct {
	PListDir string
}

func New() *Launchd {
	return &Launchd{
		PListDir: "/Library/LaunchDaemons",
	}
}

func (l *Launchd) AddDaemon(spec DaemonSpec, executable string) error {
	if _, err := os.Stat(filepath.Dir(spec.Program)); err != nil {
		os.MkdirAll(filepath.Dir(spec.Program), 0666)
	}

	if err := l.copyExecutable(executable, spec.Program); err != nil {
		return err
	}
	plistPath := filepath.Join(l.PListDir, spec.Label+".plist")
	if err := l.writePlist(spec, plistPath); err != nil {
		return err
	}
	return l.load(plistPath)
}

func (l *Launchd) RemoveDaemon(spec DaemonSpec) error {
	plistPath := filepath.Join(l.PListDir, spec.Label+".plist")
	if err := l.unload(plistPath); err != nil {
		return err
	}
	if err := os.Remove(plistPath); err != nil {
		return err
	}
	return os.Remove(spec.Program)
}

func (l *Launchd) load(plistPath string) error {
	cmd := exec.Command("launchctl", "load", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) unload(plistPath string) error {
	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l *Launchd) copyExecutable(src string, dest string) error {
	target, err := os.Create(dest)
	if err != nil {
		return err
	}

	if err = os.Chmod(dest, 0744); err != nil {
		return err
	}

	binData, err := os.Open(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(target, binData)
	return err
}

func (l *Launchd) writePlist(spec DaemonSpec, dest string) error {
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
