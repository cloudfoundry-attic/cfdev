package iso

import (
	"code.cloudfoundry.org/cfdev/provision"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Reader struct{}

func New() *Reader {
	return &Reader{}
}

type Version struct {
	Name  string `yaml:"name"`
	Value string `yaml:"version"`
}

type Metadata struct {
	Version          string              `yaml:"compatibility_version"`
	Message          string              `yaml:"splash_message"`
	DeploymentName   string              `yaml:"deployment_name"`
	AnalyticsMessage string              `yaml:"analytics_message"`
	DefaultMemory    int                 `yaml:"default_memory"`
	Services         []provision.Service `yaml:"services"`
	Versions         []Version           `yaml:"versions"`
}

func (Reader) Read(metaDataPath string) (Metadata, error) {
	buf, err := ioutil.ReadFile(metaDataPath)
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
