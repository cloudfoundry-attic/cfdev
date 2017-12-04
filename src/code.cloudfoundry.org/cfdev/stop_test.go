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
		cfdevHome   string
		hyperkitPid string
	)

	It("stops the linuxkit process", func() {
		//start up
		cfdevHome = createTempCFDevHomeDir()
		targetISOPath := filepath.Join(cfdevHome, "cfdev-efi.iso")
		hyperkitPid = filepath.Join(cfdevHome, "state", "hyperkit.pid")

		copyFileTo("./fixtures/test-image-efi.iso", targetISOPath)
		Expect(targetISOPath).To(BeAnExistingFile())

		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		Eventually(hyperkitPid).Should(BeAnExistingFile())
		eventuallyShouldListenAt("localhost:7777")

		//PID
		pidBytes, _ := ioutil.ReadFile(hyperkitPid)
		pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)

		//setup
		command = exec.Command(cliPath, "stop")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		Eventually(hyperkitPid, 30, 1).ShouldNot(BeAnExistingFile())

		//ensure pid is not running
		Expect(processIsRunning(int(pid))).To(BeFalse())
	})
})

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
