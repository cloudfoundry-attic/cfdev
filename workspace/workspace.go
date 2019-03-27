package workspace

import (
	"code.cloudfoundry.org/cfdev/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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