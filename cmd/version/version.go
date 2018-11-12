package version

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/metadata"
	"code.cloudfoundry.org/cfdev/semver"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type UI interface {
	Say(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/metadata.go code.cloudfoundry.org/cfdev/cmd/start MetaDataReader
type MetaDataReader interface {
	Read(metadataPath string) (metadata.Metadata, error)
}

type Version struct {
	UI             UI
	Version        *semver.Version
	Config         config.Config
	MetaDataReader MetaDataReader
}

func (v *Version) Execute() {
	var (
		message     = []string{fmt.Sprintf("CLI: %s", v.Version.Original)}
		metadataYml = filepath.Join(v.Config.CacheDir, "metadata.yml")
	)

	if !exists(metadataYml) {
		v.UI.Say(strings.Join(message, "\n"))
		return
	}

	metadata, err := v.MetaDataReader.Read(metadataYml)
	if err != nil {
		v.UI.Say(strings.Join(message, "\n"))
		return
	}

	for _, version := range metadata.Versions {
		message = append(message, fmt.Sprintf("%s: %s", version.Name, version.Value))
	}

	v.UI.Say(strings.Join(message, "\n"))
}

func (v *Version) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			v.Execute()
		},
	}
	return cmd
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		return false
	}

	return true
}
