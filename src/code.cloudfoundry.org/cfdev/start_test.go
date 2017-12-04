package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"github.com/onsi/gomega/gexec"
)

// TODO - hyperkit.pid is present & linuxkit cannot start
var _ = Describe("start", func() {

	var (
		cfdevHome   string
		hyperkitPid string
	)

	BeforeEach(func() {
		cfdevHome = createTempCFDevHomeDir()
		targetISOPath := filepath.Join(cfdevHome, "cfdev-efi.iso")
		hyperkitPid = filepath.Join(cfdevHome, "state", "hyperkit.pid")

		copyFileTo("./fixtures/test-image-efi.iso", targetISOPath)
		Expect(targetISOPath).To(BeAnExistingFile())
	})

	AfterEach(func() {
		pidBytes, _ := ioutil.ReadFile(hyperkitPid)
		pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

		if pid != 0 {
			syscall.Kill(int(pid), syscall.SIGTERM)
		}
	})

	It("starts the linuxkit process", func() {
		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(hyperkitPid).Should(BeAnExistingFile())
		eventuallyShouldListenAt("localhost:7777")
	})

	Context("when CFDEV_HOME is not writable", func() {
		BeforeEach(func() {
			os.Chmod(cfdevHome, 0555)
		})

		It("fails to start linuxkit", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(hyperkitPid).ShouldNot(BeAnExistingFile())
		})
	})
})

func createTempCFDevHomeDir() string {
	path, err := ioutil.TempDir("", "cfdev-home")
	Expect(err).ToNot(HaveOccurred())
	return path
}

func copyFileTo(src, dst string) {
	srcFile, err := os.Open(src)
	Expect(err).ToNot(HaveOccurred())
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0777)
	Expect(err).ToNot(HaveOccurred())
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	Expect(err).ToNot(HaveOccurred())
}

func eventuallyShouldListenAt(addr string) {
	EventuallyWithOffset(1, func() error {
		_, err := net.Dial("tcp", addr)
		return err
	}, 30).ShouldNot(HaveOccurred())
}
