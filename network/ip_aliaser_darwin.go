package network

import "fmt"

const loopback = "lo0"

func (h *HostNet) RemoveLoopbackAliases(addrs ...string) error {
	_, err := h.CfdevdClient.RemoveIPAlias()
	if err != nil {
		return err
	}

	return nil
}

func (h *HostNet) AddLoopbackAliases(addrs ...string) error {
	fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")
	_, err := h.CfdevdClient.AddIPAlias()
	if err != nil {
		return err
	}

	return nil
}
