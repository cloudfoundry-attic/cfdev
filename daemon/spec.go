package daemon

type DaemonSpec struct {
	Label                string
	EnvironmentVariables map[string]string
	Program              string
	ProgramArguments     []string
	SessionType          string
	RunAtLoad            bool
	Sockets              map[string]string
	StdoutPath           string
	StderrPath           string
	LogPath              string
	Options              map[string]interface{}
}
