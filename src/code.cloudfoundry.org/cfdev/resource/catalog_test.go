package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/resource"
)

var _ = Describe("Catalog", func() {
	var catalog resource.Catalog
	BeforeEach(func() {
		catalog = resource.Catalog{
			Items: []resource.Item{
				{
					Name: "first-resource",
					URL:  "first-resource-url",
					MD5:  "1234",
				},
				{
					Name: "second-resource",
					URL:  "second-resource-url",
					MD5:  "5678",
				},
				{
					Name: "third-resource",
					URL:  "third-resource-url",
					MD5:  "abcd",
				},
			},
		}
	})

	Describe("Lookup", func() {
		Context("the name exists", func() {
			It("returns the item", func() {
				item := catalog.Lookup("second-resource")
				Expect(item).ToNot(BeNil())
				Expect(item.MD5).To(Equal("5678"))
			})
			It("returns an item that can be edited", func() {
				item := catalog.Lookup("second-resource")
				item.MD5 = "beef"
				Expect(catalog.Lookup("second-resource").MD5).To(Equal("beef"))
			})
		})
		Context("when the name is missing", func() {
			It("returns nil", func() {
				item := catalog.Lookup("missing-resource")
				Expect(item).To(BeNil())
			})
		})
	})
})
