package cmd

import (
	"encoding/json"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

func NewCatalog(UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			bytes, err := json.MarshalIndent(Config.Dependencies, "", "  ")
			if err != nil {
				return errors.SafeWrap(err, "unable to marshal catalog")
			}
			UI.Say(string(bytes))
			return nil
		},
	}
	return cmd
}
