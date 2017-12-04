package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

		copyGardenISOTo(targetISOPath)
		Expect(targetISOPath).To(BeAnExistingFile())
	})

	AfterEach(func() {
		pidBytes, _ := ioutil.ReadFile(hyperkitPid)
		pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

		if pid != 0 {
			syscall.Kill(int(pid), syscall.SIGTERM)
		}

		os.RemoveAll(cfdevHome)
	})

	It("starts the linuxkit process", func() {
		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(hyperkitPid, 30, 1).Should(BeAnExistingFile())
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

func copyGardenISOTo(dst string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	src := filepath.Join(gopaths[0], "linuxkit", "garden-efi.iso")

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
