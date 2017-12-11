package main_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit lifecycle acceptance", func() {
	var (
		cfdevHome       string
		linuxkitPidPath string
		stateDir        string
		cacheDir        string
	)

	BeforeEach(func() {
		cfdevHome = createTempCFDevHomeDir()
		cacheDir = filepath.Join(cfdevHome, "cache")
		stateDir = filepath.Join(cfdevHome, "state")
		linuxkitPidPath = filepath.Join(stateDir, "linuxkit.pid")

		setupDependencies(cacheDir)
	})

	AfterEach(func() {
		gexec.KillAndWait()
		pid := pidFromFile("linuxkit.pid")

		if pid != 0 {
			syscall.Kill(int(-pid), syscall.SIGKILL)
		}

		os.RemoveAll(cfdevHome)
	})

	It("starts and stops the vm", func() {
		command := exec.Command(cliPath, "start")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_SKIP_ASSET_CHECK=true"),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).ShouldNot(HaveOccurred())
		Eventually(linuxkitPidPath, 10, 1).Should(BeAnExistingFile())

		// FYI - this will take time until we use thin provisioned disks
		hyperkitPidPath := filepath.Join(stateDir, "hyperkit.pid")
		Eventually(hyperkitPidPath, 120, 1).Should(BeAnExistingFile())

		By("waiting for garden to listen")
		eventuallyShouldListenAt("http://localhost:7777", 30)

		By("waiting for bosh to listen")
		eventuallyShouldListenAt("https://10.245.0.2:25555", 240)

		By("waiting for cf router to listen")
		eventuallyShouldListenAt("http://10.244.0.34:80", 1200)

		By("waiting for cfdev cli to exit when the deploy finished")
		Eventually(session, 300).Should(gexec.Exit(0))

		linuxkitPid := pidFromFile(linuxkitPidPath)
		hyperkitPid := pidFromFile(hyperkitPidPath)

		By("deploy finished - stopping...")
		command = exec.Command(cliPath, "stop")
		command.Env = append(os.Environ(),
			fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		//ensure pid is not running
		eventuallyProcessStops(linuxkitPid)
		eventuallyProcessStops(hyperkitPid)
	})
})
