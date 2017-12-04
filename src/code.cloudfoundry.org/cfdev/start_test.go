package main_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("start", func() {

	var (
		cfdevHome   string
		linuxkitPid string
		stateDir    string
	)

	BeforeEach(func() {
		cfdevHome = createTempCFDevHomeDir()
		stateDir = filepath.Join(cfdevHome, "state")
		linuxkitPid = filepath.Join(stateDir, "linuxkit.pid")

		setupDependencies(cfdevHome)
	})

	AfterEach(func() {
		gexec.KillAndWait()

		pid := pidFromFile(stateDir, "linuxkit.pid")

		if pid != 0 {
			syscall.Kill(int(-pid), syscall.SIGKILL)
		}

		os.RemoveAll(cfdevHome)
	})

	It("starts the linuxkit process", func() {
		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ShouldNot(HaveOccurred())
		Eventually(linuxkitPid, 10, 1).Should(BeAnExistingFile())
		Eventually(session, 300, 1).Should(gexec.Exit(0))

		hyperkitPid := filepath.Join(stateDir, "hyperkit.pid")
		Expect(hyperkitPid).Should(BeAnExistingFile())

		// Garden is listening
		expectToListenAt("localhost:7777")

		// BOSH Director is listening
		expectToListenAt("localhost:25555")
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
			Expect(linuxkitPid).ShouldNot(BeAnExistingFile())
		})
	})

	Context("when the state directory has existing files", func() {
		var dirtyFile string

		BeforeEach(func() {
			err := os.MkdirAll(stateDir, 0777)
			Expect(err).ToNot(HaveOccurred())

			dirtyFile = filepath.Join(stateDir, "dirty")
			err = ioutil.WriteFile(dirtyFile, []byte{}, 0777)
			Expect(err).ToNot(HaveOccurred())
		})

		It("recreates a clean state directory", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			_, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(dirtyFile, 10, 1).ShouldNot(BeAnExistingFile())
		})
	})
})

func createTempCFDevHomeDir() string {
	path, err := ioutil.TempDir("", "cfdev-home")
	Expect(err).ToNot(HaveOccurred())
	return path
}

func setupDependencies(homeDir string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	vmISO := filepath.Join(gopaths[0], "linuxkit", "cfdev-efi.iso")
	boshISO := filepath.Join(gopaths[0], "linuxkit", "bosh-deps.iso")

	targetVMPath := filepath.Join(homeDir, "cfdev-efi.iso")
	targetBoshPath := filepath.Join(homeDir, "bosh-deps.iso")

	Expect(os.Symlink(vmISO, targetVMPath)).ToNot(HaveOccurred())
	Expect(os.Symlink(boshISO, targetBoshPath)).ToNot(HaveOccurred())
}

func expectToListenAt(addr string) {
	_, err := net.Dial("tcp", addr)
	Expect(err).ShouldNot(HaveOccurred())
}
