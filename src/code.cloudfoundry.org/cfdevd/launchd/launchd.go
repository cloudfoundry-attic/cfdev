package launchd

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