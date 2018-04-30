package cmd

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

func NewCatalog(UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			bytes, err := json.MarshalIndent(Config.Dependencies, "", "  ")
			if err != nil {
				return fmt.Errorf("unable to marshal catalog: %v\n", err)
			}
			UI.Say(string(bytes))
			return nil
		},
	}
	return cmd
}
