package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/cmd"
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
		})
		Context("when the name is missing", func() {
			It("returns nil", func() {
				item := catalog.Lookup("missing-resource")
				Expect(item).To(BeNil())
			})
		})
	})

	Describe("UpdateCatalog", func(){
		originalCatalog := resource.Catalog{
			Items: []resource.Item{
				{
					Name: "original-iso-path",
					URL:  "url",
					MD5:  "md5",
					Type: "deps-iso",
					InUse: true,
				},
				{
					Name: "path to something",
					URL:  "url",
					MD5:  "md5",
					Type: "something",
					InUse: true,
				},
			},
		}

		expectedCatalog := resource.Catalog{
			Items: []resource.Item{
				{
					Name: "original-iso-path",
					URL:  "url",
					MD5:  "md5",
					Type: "deps-iso",
					InUse: false,
				},
				{
					Name: "path to something",
					URL:  "url",
					MD5:  "md5",
					Type: "something",
					InUse: true,
				},
			},
		}

		It("updates InUse flags", func(){
			args := map[string]string{
				"file": "new iso path URL",
			}

			cmd.UpdateCatalog(args, originalCatalog.Items)

			Expect(originalCatalog).Should(Equal(expectedCatalog))
		})
	})
})
