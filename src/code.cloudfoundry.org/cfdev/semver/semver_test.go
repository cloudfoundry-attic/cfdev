package semver_test

import (
	"code.cloudfoundry.org/cfdev/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semver", func() {
	It("parses 1.2.3", func() {
		s, err := semver.New("1.2.3")
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Major).To(Equal(1))
		Expect(s.Minor).To(Equal(2))
		Expect(s.Build).To(Equal(3))
		Expect(s.Original).To(Equal("1.2.3"))
	})

	It("parses 2.3.4-patch-1", func() {
		s, err := semver.New("2.3.4-patch-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Major).To(Equal(2))
		Expect(s.Minor).To(Equal(3))
		Expect(s.Build).To(Equal(4))
		Expect(s.Original).To(Equal("2.3.4-patch-1"))
	})

	It("parses empty string", func() {
		s, err := semver.New("")
		Expect(err).NotTo(HaveOccurred())
		Expect(s.Major).To(Equal(0))
		Expect(s.Minor).To(Equal(0))
		Expect(s.Build).To(Equal(0))
		Expect(s.Original).To(Equal(""))
	})
})
