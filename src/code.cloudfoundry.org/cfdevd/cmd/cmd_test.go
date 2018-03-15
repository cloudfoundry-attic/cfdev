package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"bytes"
	"code.cloudfoundry.org/cfdevd/cmd"
)

var _ = Describe("cmd", func() {
	It("return an UnimplementedCommand when given a 1", func() {
		badMessage := bytes.NewReader([]byte{uint8(7)})

		badCommand, err := cmd.UnmarshalCommand(badMessage)

		Expect(err).NotTo(HaveOccurred())
		switch v := badCommand.(type) {
		case *cmd.UnimplementedCommand:
			Expect(v.Instruction).To(Equal(uint8(7)))
		default:
			Fail("wrong type!")
		}
	})

	It("returns an BindCommand when given a 6", func() {
		badMessage := bytes.NewReader([]byte{uint8(6)})

		badCommand, err := cmd.UnmarshalCommand(badMessage)

		Expect(err).NotTo(HaveOccurred())
		switch badCommand.(type) {
		case *cmd.BindCommand:
		default:
			Fail("wrong type!")
		}
	})

	It("returns an UninstallCommand when given a 1", func() {
		badMessage := bytes.NewReader([]byte{uint8(1)})

		badCommand, err := cmd.UnmarshalCommand(badMessage)

		Expect(err).NotTo(HaveOccurred())
		switch badCommand.(type) {
		case *cmd.UninstallCommand:
		default:
			Fail("wrong type!")
		}
	})
})
