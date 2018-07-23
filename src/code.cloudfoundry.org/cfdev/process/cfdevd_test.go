// +build darwin

package process_test

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/process"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cfdevd", func() {
	Describe("IsCFDevDInstalled", func() {
		var (
			bin     string
			binDir  string
			sock    string
			sockDir string
		)

		BeforeEach(func() {
			var err error
			binDir, err = ioutil.TempDir("", "bin")
			Expect(err).NotTo(HaveOccurred())
			bin = filepath.Join(binDir, "some-bin")
			Expect(ioutil.WriteFile(bin, []byte("contents"), 0744)).To(Succeed())

			sockDir, err = ioutil.TempDir("", "sock")
			Expect(err).NotTo(HaveOccurred())
			sock = filepath.Join(sockDir, "some.socket")
		})

		AfterEach(func() {
			os.RemoveAll(sockDir)
			os.RemoveAll(binDir)
		})

		Context("installed cfdevd md5 does not match config", func() {
			It("returns false", func() {
				Expect(process.IsCFDevDInstalled(sock, bin, "bad-md5")).To(Equal(false))
			})
		})

		Context("cfdevd is not installed", func() {
			BeforeEach(func() {
				Expect(os.Remove(bin)).To(Succeed())
			})
			It("returns false", func() {
				Expect(process.IsCFDevDInstalled(sock, bin, "an-md5")).To(Equal(false))
			})
		})

		Context("installed cfdevd md5 matches config", func() {
			It("returns true if cfdevd is listening", func() {
				md5 := "98bf7d8c15784f0a3d63204441e1e2aa"
				listener := listen(sock)
				defer listener.Close()
				go accept(listener)
				Expect(process.IsCFDevDInstalled(sock, bin, md5)).To(Equal(true))
			})

			It("returns false if cfdevd is not listening", func() {
				md5 := "98bf7d8c15784f0a3d63204441e1e2aa"
				Expect(process.IsCFDevDInstalled(sock, bin, md5)).To(Equal(false))
			})
		})
	})
})

func listen(sockPath string) *net.UnixListener {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{
		Net:  "unix",
		Name: sockPath,
	})
	Expect(err).NotTo(HaveOccurred())
	return listener
}

func accept(listener *net.UnixListener) {
	defer GinkgoRecover()
	conn, err := listener.Accept()
	if err == nil {
		conn.Close()
	}
}
