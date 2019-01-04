package provision_test

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/provision"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("When progress whitelist is called with", func() {
	var (
		c        *provision.Controller
		services []provision.Service
	)

	BeforeEach(func() {
		c = provision.NewController(config.Config{})

		services = []provision.Service{
			{
				Name:          "service-one",
				Flagname:      "service-one-flagname",
				Handle:        "service-one-handle",
				Script:        "/path/to/some-script",
				Deployment:    "some-deployment",
			},
			{
				Name:          "service-two",
				Flagname:      "service-two-flagname",
				Handle:        "service-two-handle",
				Script:        "/path/to/some-script",
				Deployment:    "some-deployment",
			},
			{
				Name:          "service-three",
				Flagname:      "service-three-flagname",
				Handle:        "service-three-handle",
				Script:        "/path/to/some-script",
				Deployment:    "some-deployment",
			},
			{
				Name:          "service-four",
				Flagname:      "always-include",
				Handle:        "service-four-handle",
				Script:        "/path/to/some-script",
				Deployment:    "some-deployment",
			},
		}
	})

	Context("an empty string", func() {
		It("returns only the DefaultDeploy services", func() {
			output, err := c.WhiteListServices("", services)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(output)).To(Equal(1))
			Expect(output[0].Name).To(Equal("service-four"))
		})
	})

	Context("all", func() {
		It("returns all services", func() {
			output, err := c.WhiteListServices("all", services)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(output)).To(Equal(4))
		})
	})

	Context("service-three", func() {
		It("returns service-three and the always-include service", func() {
			output, err := c.WhiteListServices("service-three-flagname", services)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(output)).To(Equal(2))
			Expect(output[0].Name).To(Equal("service-four"))
			Expect(output[1].Name).To(Equal("service-three"))
		})
	})

	Context("multiple services", func() {
		It("returns all the requested services", func() {
			output, err := c.WhiteListServices("service-two-flagname,service-three-flagname", services)
			Expect(err).ToNot(HaveOccurred())

			Expect(len(output)).To(Equal(3))
			Expect(output[0].Name).To(Equal("service-four"))
			Expect(output[1].Name).To(Equal("service-two"))
			Expect(output[2].Name).To(Equal("service-three"))
		})
	})

	Context("none", func() {
		It("returns only the always-include service", func() {
			output, err := c.WhiteListServices("none", services)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(output)).To(Equal(1))
			Expect(output[0].Name).To(Equal("service-four"))
		})
	})
})
