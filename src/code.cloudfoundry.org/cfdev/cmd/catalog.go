package cmd

import (
	"fmt"
	"encoding/json"
	"code.cloudfoundry.org/cfdev/resource"
	"os"
)

type Catalog struct{
	UI UI
}

func (c *Catalog) Run(args []string) error {
	catalog, err := catalog(c.UI)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal catalog: %v\n", err)
	}

	c.UI.Say(string(bytes))
	return nil
}

func catalog(ui UI) (*resource.Catalog, error) {
	override := os.Getenv("CFDEV_CATALOG")

	if override != "" {
		var c resource.Catalog
		if err := json.Unmarshal([]byte(override), &c); err != nil {
			return nil, fmt.Errorf("Unable to parse CFDEV_CATALOG env variable: %v\n", err)
		}

		ui.Say("Using CFDEV_CATALOG override")
		return &c, nil
	}

	return &resource.Catalog{
		Items: []resource.Item{
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/cf-oss-deps.iso",
				Name: "cf-oss-deps.iso",
				MD5:  "c79863e02b0ee9f984c0dd5d863d6af2",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/cfdev-efi.iso",
				Name: "cfdev-efi.iso",
				MD5:  "fd1e13bb7badcacefc4e810d12a83b1d",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/vpnkit",
				Name: "vpnkit",
				MD5:  "4eb4c3477e8296f4e97b5c89983d4ff3",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/hyperkit",
				Name: "hyperkit",
				MD5:  "61da21b4e82e2bf2e752d043482aa966",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/linuxkit",
				Name: "linuxkit",
				MD5:  "9ae23eec8d297f41caff3450d6a03b3c",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/qcow-tool",
				Name: "qcow-tool",
				MD5:  "22f3a57096ae69027c13c4933ccdd96c",
			},
			{
				URL:  "https://s3.amazonaws.com/pcfdev-development/stories/153571042/UEFI.fd",
				Name: "UEFI.fd",
				MD5:  "2eff1c02d76fc3bde60f497ce1116b09",
			},
		},
	}, nil
}
