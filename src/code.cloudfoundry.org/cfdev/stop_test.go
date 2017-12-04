package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("stop", func() {
	var (
		cfdevHome string
		stateDir  string
	)

	BeforeEach(func() {
		cfdevHome = createTempCFDevHomeDir()
		stateDir = filepath.Join(cfdevHome, "state")

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

	It("stops the linuxkit process", func() {
		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, 300, 1).Should(gexec.Exit(0))

		// All services are up
		expectToListenAt("localhost:7777")
		expectToListenAt("localhost:25555")

		//PID
		linuxkitPid := pidFromFile(stateDir, "linuxkit.pid")
		hyperkitPid := pidFromFile(stateDir, "hyperkit.pid")

		//setup
		command = exec.Command(cliPath, "stop")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		//ensure pid is not running
		Eventually(func() (bool, error) {
			return processIsRunning(linuxkitPid)
		}).Should(BeFalse())

		Eventually(func() (bool, error) {
			return processIsRunning(hyperkitPid)
		}).Should(BeFalse())
	})
})

func pidFromFile(stateDir, pidFile string) int {
	path := filepath.Join(stateDir, pidFile)
	pidBytes, _ := ioutil.ReadFile(path)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)
	return int(pid)
}

func processIsRunning(pid int) (bool, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}
