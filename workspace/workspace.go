package workspace

import (
	"archive/tar"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"compress/gzip"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Version struct {
	Name  string `yaml:"name"`
	Value string `yaml:"version"`
}

type Service struct {
	Name       string `yaml:"name"`
	Flagname   string `yaml:"flag_name"`
	Script     string `yaml:"script"`
	Deployment string `yaml:"deployment"`
	IsErrand   bool   `yaml:"errand"`
}

type Metadata struct {
	Version          string              `yaml:"compatibility_version"`
	ArtifactVersion  string              `yaml:"artifact_version"`
	Message          string              `yaml:"splash_message"`
	DeploymentName   string              `yaml:"deployment_name"`
	AnalyticsMessage string              `yaml:"analytics_message"`
	DefaultMemory    int                 `yaml:"default_memory"`
	Services         []Service `yaml:"services"`
	Versions         []Version           `yaml:"versions"`
}

type Workspace struct {
	Config config.Config
}

func New(config config.Config) *Workspace {
	return &Workspace{
		Config: config,
	}
}

func (w *Workspace) CreateDirs() error {
	err := removeDirAlls(
		w.Config.LogDir,
		w.Config.StateDir,
		w.Config.BinaryDir,
		w.Config.ServicesDir,
		w.Config.DaemonDir)
	if err != nil {
		return err
	}

	return mkdirAlls(
		w.Config.CacheDir,
		w.Config.DaemonDir,
		w.Config.LogDir)
}

func (w *Workspace) SetupState(depsFile string) error {
	f, err := os.Open(depsFile)
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

		target := filepath.Join(w.Config.CFDevHome, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}

	return nil
}

func (w *Workspace) EnvsMapping() map[string]string {
	mapping := map[string]string{}

	data, err := ioutil.ReadFile(filepath.Join(w.Config.StateBosh, "env.yml"))
	if err != nil {
		return mapping
	}

	yaml.Unmarshal(data, &mapping)
	return mapping
}

func (w *Workspace) Envs() []string {
	var results []string
	for k, v := range w.EnvsMapping() {
		results = append(results, k+"="+v)
	}

	return results
}

func (w *Workspace) Metadata() (Metadata, error) {
	buf, err := ioutil.ReadFile(filepath.Join(w.Config.StateDir, "metadata.yml"))
	if err != nil {
		return Metadata{}, err
	}

	var metadata Metadata

	err = yaml.Unmarshal(buf, &metadata)
	if err != nil {
		return Metadata{}, err
	}

	return metadata, nil
}

func mkdirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to create dir")
		}
	}

	return nil
}

func removeDirAlls(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(fmt.Errorf("path %s: %s", dir, err), "failed to remove dir")
		}
	}

	return nil
}