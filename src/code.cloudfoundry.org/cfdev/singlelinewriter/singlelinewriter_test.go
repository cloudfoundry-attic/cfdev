package singlelinewriter_test

import (
	"bytes"

	"code.cloudfoundry.org/cfdev/singlelinewriter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SingleLineWriter", func() {
	var (
		buf     bytes.Buffer
		subject singlelinewriter.UI
	)
	BeforeEach(func() {
		buf.Reset()
		subject = singlelinewriter.New(&buf)
	})

	It("Sends ansi save upon creation", func() {
		Eventually(buf.Bytes).Should(Equal([]byte("\033[s\033[7l")))
	})

	Describe("Say", func() {
		It("writes to writer", func() {
			subject.Say("Hi Mom")
			Eventually(buf.String).Should(ContainSubstring("Hi Mom"))
		})
	})

	Describe("Write", func() {
		It("writes to writer", func() {
			Expect(subject.Write([]byte("Hi Mom\nBye Dad\n"))).To(Equal(15))
			Eventually(buf.String).Should(ContainSubstring("Hi Mom"))
			Eventually(buf.String).Should(ContainSubstring("Bye Dad"))
		})
		It("Sends ansi restore & delete to end commands", func() {
			Expect(subject.Write([]byte("Hi Mom\n"))).To(Equal(7))
			Eventually(buf.Bytes).Should(ContainSubstring("\033[u\033[J"))
		})
	})
	Describe("Close", func() {
		It("Sends ansi restore & delete to end commands", func() {
			Expect(subject.Close()).To(Succeed())
			Eventually(buf.Bytes).Should(Equal([]byte("\033[s\033[7l\033[u\033[J")))
		})
	})
})
