package util_test

import (
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {

	var tmpDir string

	BeforeEach(func() {
		tmpDir, _ = ioutil.TempDir("", "cfdev.util.")
		_ = ioutil.WriteFile(filepath.Join(tmpDir, "dat1"), []byte("contents"), 0644)
	})

	It("copies file", func() {
		_ = util.CopyFile(filepath.Join(tmpDir, "dat1"), filepath.Join(tmpDir, "dat2"))
		Expect(filepath.Join(tmpDir, "dat2")).To(BeAnExistingFile())
		Expect(ioutil.ReadFile(filepath.Join(tmpDir, "dat2"))).To(Equal([]byte("contents")))
	})
})
