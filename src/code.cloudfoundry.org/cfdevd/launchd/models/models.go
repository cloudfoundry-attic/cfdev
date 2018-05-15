package models

type DaemonSpec struct {
	Label            string
	Program          string
	ProgramArguments []string
	SessionType      string
	RunAtLoad        bool
	Sockets          map[string]string
	StdoutPath       string
	StderrPath       string
}
