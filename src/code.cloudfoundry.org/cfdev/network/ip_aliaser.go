package network

import (
	"fmt"
)

type HostNet struct{}

func (*HostNet) AddLoopbackAliases(addrs ...string) error {
	fmt.Println("Setting up IP aliases for the BOSH Director & CF Router (requires administrator privileges)")

	if err := createInterface(); err != nil {
		return err
	}

	for _, addr := range addrs {
		exists, err := aliasExists(addr)

		if err != nil {
			return err
		}

		if exists {
			continue
		}

		err = addAlias(addr)
		if err != nil {
			return err
		}
	}
	return nil
}
