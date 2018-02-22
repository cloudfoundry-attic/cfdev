package cmd

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cfdev/config"
)

type Catalog struct {
	UI     UI
	Config config.Config
}

func (c *Catalog) Run(args []string) error {
	bytes, err := json.MarshalIndent(c.Config.Dependencies, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal catalog: %v\n", err)
	}

	c.UI.Say(string(bytes))
	return nil
}
