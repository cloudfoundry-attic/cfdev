package host

//go:generate mockgen -package mocks -destination mocks/powershell.go code.cloudfoundry.org/cfdev/host Powershell
type Powershell interface {
	Output(command string) (string, error)
}

type Host struct {
	Powershell Powershell
}
