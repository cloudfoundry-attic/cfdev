package version

import (
	"archive/tar"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/metadata"
	"code.cloudfoundry.org/cfdev/semver"
	"compress/gzip"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
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
	BuildVersion   string
	Config         config.Config
	MetaDataReader MetaDataReader
}

func (v *Version) Execute(pathTarball string) {
	var (
		message     []string
		tmpDir      string
		metadataYml = filepath.Join(v.Config.StateDir, "metadata.yml")
	)

	if pathTarball != "" {
		if !exists(pathTarball) {
			v.UI.Say(fmt.Sprintf("%s: file not found", pathTarball))
			return
		}

		tmpDir, _ = ioutil.TempDir("", "versioncmd")
		defer os.RemoveAll(tmpDir)

		untarFile(pathTarball, tmpDir, "metadata.yml")

		if !exists(filepath.Join(tmpDir, "metadata.yml")) {
			v.UI.Say("Metadata not found version unknown")
			return
		}

		metadataYml = filepath.Join(tmpDir, "metadata.yml")
	}

	v.printCliVersion()

	if exists(metadataYml) {
		mtData, err := v.MetaDataReader.Read(metadataYml)
		if err != nil {
			return
		}

		for _, version := range mtData.Versions {
			message = append(message, fmt.Sprintf("%s: %s", version.Name, version.Value))
		}

		v.UI.Say(strings.Join(message, "\n"))
	}
}

func (v *Version) printCliVersion() {
	v.UI.Say(fmt.Sprintf("CLI: %s\nBUILD: %s\n", v.Version.Original, v.BuildVersion))
}

func (v *Version) Cmd() *cobra.Command {
	filename := ""

	cmd := &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			v.Execute(filename)
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&filename, "file", "f", "", "path to deps-tar file")
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

func untarFile(tarballPath, destinationDir, name string) error {
	f, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		if filepath.Base(header.Name) == name {
			target := filepath.Join(destinationDir, filepath.Base(header.Name))

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
			return nil
		}
	}

	return nil
}
