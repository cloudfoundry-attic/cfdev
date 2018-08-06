package iso

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/garden"
	"github.com/hooklift/iso9660"
	yaml "gopkg.in/yaml.v2"
)

type Reader struct{}

func New() *Reader {
	return &Reader{}
}

type Metadata struct {
	Version       string           `yaml:"compatibility_version"`
	Message       string           `yaml:"splash_message"`
	DefaultMemory int              `yaml:"default_memory"`
	Services      []garden.Service `yaml:"services"`
}

func (Reader) Read(isoFile string) (Metadata, error) {
	file, err := os.Open(isoFile)
	if err != nil {
		return Metadata{}, err
	}

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
