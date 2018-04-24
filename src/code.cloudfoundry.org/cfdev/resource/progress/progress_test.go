package progress_test

import (
	"bytes"
	"strings"

	"code.cloudfoundry.org/cfdev/resource/progress"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer Progress", func() {
	Describe("#Read", func() {
		var (
			stdout  bytes.Buffer
			subject *progress.Progress
		)

		BeforeEach(func() {
			stdout = bytes.Buffer{}
			subject = progress.New(&stdout)
		})

		It("displays 0%", func() {
			subject.Start(1000)
			subject.Write([]byte{})
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |>                    | 0.0%"))
		})

		It("writing all bytes displays 100%", func() {
			subject.Start(1000)
			subject.Write(bytes.Repeat([]byte(" "), 1000))
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |====================>| 100.0%"))
		})

		It("adding all bytes displays 100%", func() {
			subject.Start(1000)
			subject.Add(1000)
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |====================>| 100.0%"))
		})

		It("sending multiple bytes displays 50%", func() {
			subject.Start(1000)
			subject.Write(bytes.Repeat([]byte(" "), 250))
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |=====>               | 25.0%"))
			subject.Add(250)
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |==========>          | 50.0%"))
		})

		It("shows percentage as one decimal place 50%", func() {
			subject.Start(1000000)
			subject.Write(bytes.Repeat([]byte(" "), 251234))
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |=====>               | 25.1%"))
			subject.Add(251234)
			Expect(stdout.String()).To(ContainSubstring("\rProgress: |==========>          | 50.2%"))
		})

		It("supresses printing the same percentage", func() {
			split := func() []string { return strings.Split(stdout.String(), "\r") }
			subject.Start(10000)
			subject.Write(bytes.Repeat([]byte(" "), 2500))
			Expect(split()).To(Equal([]string{
				"",
				"Progress: |>                    | 0%",
				"Progress: |=====>               | 25.0%",
			}))
			subject.Write(bytes.Repeat([]byte(" "), 4))
			subject.Write(bytes.Repeat([]byte(" "), 4))
			Expect(split()).To(Equal([]string{
				"",
				"Progress: |>                    | 0%",
				"Progress: |=====>               | 25.0%",
			}))
			subject.Write(bytes.Repeat([]byte(" "), 4))
			Expect(split()).To(Equal([]string{
				"",
				"Progress: |>                    | 0%",
				"Progress: |=====>               | 25.0%",
				"Progress: |=====>               | 25.1%",
			}))
		})

		It("clears the line upon calling End", func() {
			subject.Start(1000)
			subject.End()
			Expect(stdout.String()).To(ContainSubstring("\r"))
		})

		It("handles setting total to 0", func() {
			subject.Start(0)
			subject.Write(bytes.Repeat([]byte(" "), 251))
			Expect(stdout.String()).To(ContainSubstring("\rProgress: 251 bytes"))
			subject.Add(250)
			Expect(stdout.String()).To(ContainSubstring("\rProgress: 501 bytes"))
		})
	})
})
