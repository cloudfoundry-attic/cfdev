package errors_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SafeWrap", func() {
	It("returns a wrapped error", func() {
		err := errors.SafeWrap(fmt.Errorf("other"), "safe text")
		Expect(err).To(MatchError("safe text: other"))
	})

	Describe("SafeError", func() {
		It("returns ONLY the safe errors", func() {
			err := errors.SafeWrap(fmt.Errorf("other"), "safe text")
			Expect(errors.SafeError(err)).To(Equal("safe text"))
		})

		It("returns ALL the safe errors", func() {
			err := errors.SafeWrap(fmt.Errorf("other"), "safe text")
			err = errors.SafeWrap(err, "outer text")
			Expect(errors.SafeError(err)).To(Equal("outer text: safe text"))
		})

		It("returns empty string for non safe errors", func() {
			err := fmt.Errorf("other")
			Expect(errors.SafeError(err)).To(Equal(""))
		})
	})
})
