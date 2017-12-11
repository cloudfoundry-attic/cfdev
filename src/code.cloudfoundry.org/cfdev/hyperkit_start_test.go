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
	"sync/atomic"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit start acceptance", func() {

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

	Context("with an unsupported distribution", func() {
		It("exits with code 1", func() {
			command := exec.Command(cliPath, "start", "-f", "UNSUPPORTTED")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))
			session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with an unsupported version", func() {
		It("exits with code 1", func() {
			command := exec.Command(cliPath, "start", "-n", "9.9.9.9.9")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))
			session, _ := gexec.Start(command, GinkgoWriter, GinkgoWriter)

			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when CFDEV_HOME is not writable", func() {
		BeforeEach(func() {
			os.Chmod(cfdevHome, 0555)
		})

		AfterEach(func() {
			os.Chmod(cfdevHome, 0777)
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

	Context("the linuxkit pid file references an existing process", func() {
		var (
			existingCmd *exec.Cmd
			existingPid int
			exited      int32
		)

		BeforeEach(func() {
			err := os.MkdirAll(stateDir, 0777)
			Expect(err).ToNot(HaveOccurred())

			existingCmd = exec.Command("sleep", "300")
			err = existingCmd.Start()
			Expect(err).ToNot(HaveOccurred())

			existingPid = existingCmd.Process.Pid
			err = ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(existingPid)), 0777)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				existingCmd.Wait()
				atomic.StoreInt32(&exited, 1)
			}()
		})

		AfterEach(func() {
			existingCmd.Process.Kill()
		})

		It("doesn't restart the linuxkit process", func() {
			command := exec.Command(cliPath, "start")
			command.Env = append(os.Environ(),
				fmt.Sprintf("CFDEV_HOME=%s", cfdevHome))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session, 10, 1).Should(gexec.Exit(0))

			Expect(pidFromFile(linuxkitPidPath)).To(Equal(existingPid))
			Expect(atomic.LoadInt32(&exited)).To(BeEquivalentTo(0))
		})
	})
})

func setupDependencies(cacheDir string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")

	assets := []string{
		"cfdev-efi.iso",
		"cf-deps.iso",
		"bosh-deps.iso",
		"vpnkit",
		"hyperkit",
		"linuxkit",
		"UEFI.fd",
	}

	err := os.MkdirAll(cacheDir, 0777)
	Expect(err).ToNot(HaveOccurred())

	for _, asset := range assets {
		origin := filepath.Join(gopaths[0], "linuxkit", asset)
		target := filepath.Join(cacheDir, asset)

		Expect(origin).To(BeAnExistingFile())
		Expect(os.Symlink(origin, target)).ToNot(HaveOccurred())
	}
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
