package iso

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/provision"
	"github.com/hooklift/iso9660"
	"gopkg.in/yaml.v2"
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
	Version        string              `yaml:"compatibility_version"`
	Message        string              `yaml:"splash_message"`
	DeploymentName string              `yaml:"deployment_name"`
	DefaultMemory  int                 `yaml:"default_memory"`
	Services       []provision.Service `yaml:"services"`
	Versions       []Version           `yaml:"versions"`
}

func (Reader) Read(isoFile string) (Metadata, error) {
	file, err := os.Open(isoFile)
	if err != nil {
		return Metadata{}, err
	}
	defer file.Close()

	r, err := iso9660.NewReader(file)
	if err != nil {
		return Metadata{}, err
	}

	for {
		f, err := r.Next()
		if err == io.EOF {
			fmt.Println("File not found")
			return Metadata{}, err
		}

		if err != nil {
			return Metadata{}, err
		}

		if strings.Contains(f.Name(), "metadata.yml") {
			buf, err := ioutil.ReadAll(f.Sys().(io.Reader))
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
	}
}
