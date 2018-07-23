package launchd

type DaemonSpec struct {
	Label            string
	CfDevHome        string
	Program          string
	ProgramArguments []string
	SessionType      string
	RunAtLoad        bool
	Sockets          map[string]string
	StdoutPath       string
	StderrPath       string
}
