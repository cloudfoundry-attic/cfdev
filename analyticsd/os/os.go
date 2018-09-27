package os

//go:generate mockgen -package mocks -destination mocks/runner.go code.cloudfoundry.org/cfdev/analyticsd/os Runner
type Runner interface{
	Output(command string, arg ...string) (output []byte, err error)
}

type OS struct {
	Runner Runner
}

//sw_vers -productVersion
