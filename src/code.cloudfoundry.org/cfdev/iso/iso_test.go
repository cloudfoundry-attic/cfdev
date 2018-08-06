package iso_test

import (
	"code.cloudfoundry.org/cfdev/iso"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Iso", func() {
	Context("reader returns", func() {
		It("metadata", func() {
			metadata, err := iso.New().Read("fixtures/cf-deps.iso")

			Expect(err).ToNot(HaveOccurred())
			Expect(metadata.Version).To(Equal("v29"))
			Expect(metadata.Message).To(Equal("is simply dummy text"))
		})
	})
})
