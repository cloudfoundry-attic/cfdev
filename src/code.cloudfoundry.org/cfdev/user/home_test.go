package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cfdev/user"
)

var _ = Describe("CFDevHome", func() {
	It("creates & returns the cfdev home directory", func() {
		path, err := user.CFDevHome()

		Expect(err).ToNot(HaveOccurred())
		Expect(path).To(HaveSuffix(".cfdev"))
		Expect(path).To(BeADirectory())
	})

	Context("when CFDEV_HOME is set", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "some-other-home")

			Expect(err).ToNot(HaveOccurred())
			os.Setenv("CFDEV_HOME", tmpDir)
		})

		AfterEach(func() {
			os.Unsetenv("CFDEV_HOME")
		})

		It("returns the override for the home directory", func() {
			path, err := user.CFDevHome()

			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(HavePrefix(tmpDir))
			Expect(path).ToNot(HaveSuffix(".cfdev"))
			Expect(path).To(BeADirectory())
		})
	})
})
