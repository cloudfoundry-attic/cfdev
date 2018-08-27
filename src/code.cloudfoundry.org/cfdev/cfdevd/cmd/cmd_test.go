// +build darwin

package cmd_test

import (
	"bytes"

	"code.cloudfoundry.org/cfdev/cfdevd/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cmd", func() {
	It("return an UnimplementedCommand when given a 7", func() {
		badMessage := bytes.NewReader([]byte{uint8(7)})

		badCommand, err := cmd.UnmarshalCommand(badMessage)

		Expect(err).NotTo(HaveOccurred())
		switch v := badCommand.(type) {
		case *cmd.UnimplementedCommand:
			Expect(v.Instruction).To(Equal(uint8(7)))
			Expect(v.Logger).ToNot(BeNil())
		default:
			Fail("wrong type!")
		}
	})

	It("returns an BindCommand when given a 6", func() {
		message := bytes.NewReader([]byte{uint8(6)})

		command, err := cmd.UnmarshalCommand(message)

		Expect(err).NotTo(HaveOccurred())
		switch command.(type) {
		case *cmd.BindCommand:
		default:
			Fail("wrong type!")
		}
	})

	It("returns an UninstallCommand when given a 1", func() {
		message := bytes.NewReader([]byte{uint8(1)})

		command, err := cmd.UnmarshalCommand(message)

		Expect(err).NotTo(HaveOccurred())
		switch command.(type) {
		case *cmd.UninstallCommand:
		default:
			Fail("wrong type!")
		}
	})

	It("returns a RemoveIPAliasCommand", func() {
		message := bytes.NewReader([]byte{uint8(2)})

		command, err := cmd.UnmarshalCommand(message)

		Expect(err).NotTo(HaveOccurred())
		switch command.(type) {
		case *cmd.RemoveIPAliasCommand:
		default:
			Fail("wrong type!")
		}
	})

	It("returns a AddIPAliasCommand", func() {
		message := bytes.NewReader([]byte{uint8(3)})

		command, err := cmd.UnmarshalCommand(message)

		Expect(err).NotTo(HaveOccurred())
		switch command.(type) {
		case *cmd.AddIPAliasCommand:
		default:
			Fail("wrong type!")
		}
	})
})
