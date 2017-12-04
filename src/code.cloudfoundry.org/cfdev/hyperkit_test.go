package main_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
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

var _ = Describe("hyperkit acceptance", func() {

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
		eventuallyShouldListenAt("https://localhost:25555", 240)

		By("waiting for cf router to listen")
		eventuallyShouldListenAt("http://localhost:35555", 1200)

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
			Expect(linuxkitPidPath).ShouldNot(BeAnExistingFile())
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

func setupDependencies(cacheDir string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	vmISO := filepath.Join(gopaths[0], "linuxkit", "cfdev-efi.iso")
	cfISO := filepath.Join(gopaths[0], "linuxkit", "cf-deps.iso")
	boshISO := filepath.Join(gopaths[0], "linuxkit", "bosh-deps.iso")

	Expect(vmISO).To(BeAnExistingFile())
	Expect(boshISO).To(BeAnExistingFile())
	Expect(cfISO).To(BeAnExistingFile())

	err := os.MkdirAll(cacheDir, 0777)
	Expect(err).ToNot(HaveOccurred())

	targetVMPath := filepath.Join(cacheDir, "cfdev-efi.iso")
	targetBoshPath := filepath.Join(cacheDir, "bosh-deps.iso")
	targetCFPath := filepath.Join(cacheDir, "cf-deps.iso")

	Expect(os.Symlink(vmISO, targetVMPath)).ToNot(HaveOccurred())
	Expect(os.Symlink(boshISO, targetBoshPath)).ToNot(HaveOccurred())
	Expect(os.Symlink(cfISO, targetCFPath)).ToNot(HaveOccurred())
}

func eventuallyShouldListenAt(url string, timeoutSec int) {
	Eventually(func() error {
		return httpServerIsListeningAt(url)
	}, timeoutSec, 1).ShouldNot(HaveOccurred())
}

func httpServerIsListeningAt(url string) error {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get(url)

	if resp != nil {
		resp.Body.Close()
	}

	return err
}

func eventuallyProcessStops(pid int) {
	EventuallyWithOffset(1, func() (bool, error) {
		return processIsRunning(pid)
	}).Should(BeFalse())
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

func pidFromFile(pidFile string) int {
	pidBytes, _ := ioutil.ReadFile(pidFile)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)
	return int(pid)
}
