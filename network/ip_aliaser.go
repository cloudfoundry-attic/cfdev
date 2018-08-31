package network

//go:generate mockgen -package mocks -destination mocks/cfdevd_client.go code.cloudfoundry.org/cfdev/network CfdevdClient
type CfdevdClient interface {
	Uninstall() (string, error)
	AddIPAlias() (string, error)
	RemoveIPAlias() (string, error)
}

type HostNet struct{
	CfdevdClient CfdevdClient
}

