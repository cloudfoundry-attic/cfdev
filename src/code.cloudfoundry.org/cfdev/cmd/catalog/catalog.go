package catalog

import (
	"encoding/json"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Catalog struct {
	UI     UI
	Config config.Config
}

func (c *Catalog) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "catalog",
		RunE: c.RunE,
	}
}

func (c *Catalog) RunE(cmd *cobra.Command, args []string) error {
	bytes, err := json.MarshalIndent(c.Config.Dependencies, "", "  ")
	if err != nil {
		return errors.SafeWrap(err, "unable to marshal catalog")
	}
	c.UI.Say(string(bytes))
	return nil
}
