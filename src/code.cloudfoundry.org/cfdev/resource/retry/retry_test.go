package retry_test

import (
	"bytes"
	"fmt"
	"time"

	"code.cloudfoundry.org/cfdev/resource/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry", func() {
	It("retries until success", func() {
		counter := 0
		fn := func() error {
			counter += 1
			if counter < 6 {
				return fmt.Errorf("failing")
			}
			return nil
		}
		retryFn := func(error) bool { return true }
		Expect(retry.Retry(fn, retryFn)).To(Succeed())
		Expect(counter).To(Equal(6))
	})

	It("returns error if retryFn returns false", func() {
		fn := func() error { return fmt.Errorf("failing") }
		retryFn := func(error) bool { return false }

		Expect(retry.Retry(fn, retryFn)).To(MatchError("failing"))
	})

	Describe("Retryable", func() {
		var buffer bytes.Buffer
		It("does not retry other errors", func() {
			counter := 0
			fn := func() error {
				counter++
				return fmt.Errorf("failing")
			}

			Expect(retry.Retry(fn, retry.Retryable(10, time.Nanosecond, &buffer))).To(MatchError("failing"))
			Expect(counter).To(Equal(1))
			Expect(buffer.String()).NotTo(ContainSubstring("Failed: Retrying:"))
		})

		It("retries retyables a max number of times", func() {
			counter := 0
			fn := func() error {
				counter++
				return retry.WrapAsRetryable(fmt.Errorf("failing"))
			}

			Expect(retry.Retry(fn, retry.Retryable(10, time.Nanosecond, &buffer))).To(MatchError("failing"))
			Expect(counter).To(Equal(10))
			Expect(buffer.String()).To(ContainSubstring("Failed: Retrying:"))
		})
	})
})
