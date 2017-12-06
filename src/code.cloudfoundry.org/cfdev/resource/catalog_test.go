package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/resource"
)

var _ = Describe("Catalog", func() {

	It("can prune items that do not target the operating system", func() {
		c := resource.Catalog{
			Items: []resource.Item{
				{Name: "first-resource", OS: "darwin"},
				{Name: "second-resource"},
				{Name: "third-resource", OS: "windows"},
			},
		}

		Expect(c.Filter("darwin")).To(Equal(&resource.Catalog{
			Items: []resource.Item{
				{Name: "first-resource", OS: "darwin"},
				{Name: "second-resource"},
			},
		}))
	})
})
