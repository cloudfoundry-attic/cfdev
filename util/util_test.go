package util_test

import (
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/util"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("Util", func() {
	Describe("CopyFile", func() {
		var tmpDir string

		BeforeEach(func() {
			tmpDir, _ = ioutil.TempDir("", "cfdev.util.")
			_ = ioutil.WriteFile(filepath.Join(tmpDir, "dat1"), []byte("contents"), 0644)
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("copies file", func() {
			_ = util.CopyFile(filepath.Join(tmpDir, "dat1"), filepath.Join(tmpDir, "dat2"))
			Expect(filepath.Join(tmpDir, "dat2")).To(BeAnExistingFile())
			Expect(ioutil.ReadFile(filepath.Join(tmpDir, "dat2"))).To(Equal([]byte("contents")))
		})
	})

	Describe("Perform", func() {

		var (
			counter int
			task    = func() error {
				if counter > 3 { return nil }

				counter++

				return errors.New("some-flake-error")
			}
		)

		BeforeEach(func() {
			counter = 0
		})

		It("retries as much as specified to complete a task", func() {
			err := util.Perform(5, task)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error if all attempts fail", func() {
			err := util.Perform(2, task)
			Expect(err).To(MatchError("some-flake-error"))
		})
	})
})