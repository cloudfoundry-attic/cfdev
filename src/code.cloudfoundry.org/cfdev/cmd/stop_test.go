package cmd_test

import (
	// . "code.cloudfoundry.org/cfdev/cmd"

	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/spf13/cobra"
)

type MockClient struct{}

func (mc *MockClient) Event(string, map[string]interface{}) error      { return nil }
func (mc *MockClient) Close()                                          {}
func (mc *MockClient) PromptOptIn(chan struct{}, cfanalytics.UI) error { return nil }

var _ = Describe("Stop", func() {
	var (
		tmpDir                                                 string
		linuxkit, vpnkit, hyperkit                             *gexec.Session
		state, linuxkitPidPath, hyperkitPidPath, vpnkitPidPath string
		cfg                                                    config.Config
		stopCmd                                                *cobra.Command
	)

	BeforeEach(func() {
		state, _ = ioutil.TempDir("", "pcfdev.stop.")

		linuxkitPidPath = filepath.Join(state, "linuxkit.pid")
		hyperkitPidPath = filepath.Join(state, "hyperkit.pid")
		vpnkitPidPath = filepath.Join(state, "vpnkit.pid")
		tmpDir, _ = ioutil.TempDir("", "cfdev.stop.")

		cfg = config.Config{
			LinuxkitPidFile:  linuxkitPidPath,
			HyperkitPidFile:  hyperkitPidPath,
			VpnkitPidFile:    vpnkitPidPath,
			Analytics:        &MockClient{},
			CFDevDSocketPath: filepath.Join(tmpDir, "cfdevd.socket"),
		}
		stopCmd = cmd.NewStop(cfg)
		stopCmd.SetArgs([]string{})
		stopCmd.SetOutput(GinkgoWriter)
	})
	AfterEach(func() {
		os.RemoveAll(state)
		os.RemoveAll(tmpDir)
	})
	Context("all processes are running and pid files exist", func() {
		BeforeEach(func() {
			linuxkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)
			vpnkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)
			hyperkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)

			ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(linuxkit.Command.Process.Pid)), 0644)
			ioutil.WriteFile(vpnkitPidPath, []byte(strconv.Itoa(vpnkit.Command.Process.Pid)), 0644)
			ioutil.WriteFile(hyperkitPidPath, []byte(strconv.Itoa(hyperkit.Command.Process.Pid)), 0644)
		})

		AfterEach(func() {
			linuxkit.Terminate()
			vpnkit.Terminate()
			hyperkit.Terminate()
		})

		It("kill all Pids", func() {
			Expect(stopCmd.Execute()).To(Succeed())

			Eventually(linuxkit).Should(gexec.Exit())
			Eventually(vpnkit).Should(gexec.Exit())
			Eventually(hyperkit).Should(gexec.Exit())
		})

		It("removes the pid files", func() {
			Expect(stopCmd.Execute()).To(Succeed())

			Expect(linuxkitPidPath).ToNot(BeAnExistingFile())
			Expect(vpnkitPidPath).ToNot(BeAnExistingFile())
			Expect(hyperkitPidPath).ToNot(BeAnExistingFile())
		})
	})

	Context("all pidfiles are missing", func() {
		It("does nothing and succeeds", func() {
			Expect(stopCmd.Execute()).To(Succeed())
		})
	})

	Context("one pid file is missing", func() {
		BeforeEach(func() {
			os.Remove(vpnkitPidPath)
		})

		It("kills existing pids", func() {
			Expect(stopCmd.Execute()).To(Succeed())

			Expect(linuxkitPidPath).ToNot(BeAnExistingFile())
			Expect(hyperkitPidPath).ToNot(BeAnExistingFile())
		})
	})

	Context("one process has stopped, pid file exists", func() {
		BeforeEach(func() {
			linuxkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)
			vpnkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)
			hyperkit, _ = gexec.Start(exec.Command("sleep", "100"), GinkgoWriter, GinkgoWriter)

			Expect(ioutil.WriteFile(linuxkitPidPath, []byte(strconv.Itoa(linuxkit.Command.Process.Pid)), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(vpnkitPidPath, []byte(strconv.Itoa(vpnkit.Command.Process.Pid)), 0644)).To(Succeed())
			Expect(ioutil.WriteFile(hyperkitPidPath, []byte(strconv.Itoa(hyperkit.Command.Process.Pid)), 0644)).To(Succeed())

			vpnkit.Kill()
			Eventually(vpnkit).Should(gexec.Exit())
		})

		It("kills existing pids and returns error", func() {
			Expect(stopCmd.Execute()).To(Succeed())

			Eventually(linuxkit).Should(gexec.Exit())
			Eventually(hyperkit).Should(gexec.Exit())

			Expect(linuxkitPidPath).ToNot(BeAnExistingFile())
			Expect(vpnkitPidPath).ToNot(BeAnExistingFile())
			Expect(hyperkitPidPath).ToNot(BeAnExistingFile())
		})
	})

	Context("all processes have exited and all pidfiles exists", func() {
		BeforeEach(func() {
			proc, _ := gexec.Start(exec.Command("echo", "100"), GinkgoWriter, GinkgoWriter)
			Eventually(proc).Should(gexec.Exit(0))
			pid := []byte(strconv.Itoa(proc.Command.Process.Pid))

			ioutil.WriteFile(linuxkitPidPath, pid, 0644)
			ioutil.WriteFile(vpnkitPidPath, pid, 0644)
			ioutil.WriteFile(hyperkitPidPath, pid, 0644)
		})

		It("deletes all pid files and succeeds", func() {
			Expect(stopCmd.Execute()).To(Succeed())

			Expect(linuxkitPidPath).ToNot(BeAnExistingFile())
			Expect(vpnkitPidPath).ToNot(BeAnExistingFile())
			Expect(hyperkitPidPath).ToNot(BeAnExistingFile())
		})
	})

	Context("cfdevd socket exists", func() {
		var instructions chan byte
		var uninstallErrorCode int
		BeforeEach(func() {
			instructions = make(chan byte, 1)
			ln, err := net.Listen("unix", cfg.CFDevDSocketPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.CFDevDSocketPath).To(BeAnExistingFile())
			go func() {
				conn, err := ln.Accept()
				Expect(err).NotTo(HaveOccurred())
				handshake := make([]byte, 49, 49)
				binary.Read(conn, binary.LittleEndian, handshake)
				binary.Write(conn, binary.LittleEndian, handshake)
				instruction := make([]byte, 1, 1)
				binary.Read(conn, binary.LittleEndian, instruction)
				instructions <- instruction[0]
				if uninstallErrorCode == -1 {
					conn.Close()
				} else {
					binary.Write(conn, binary.LittleEndian, []byte{byte(uninstallErrorCode)})
				}
			}()
		})
		It("succeeds and sends the uninstall command to cfdevd", func() {
			uninstallErrorCode = 0
			Expect(stopCmd.Execute()).To(Succeed())

			Eventually(instructions).Should(Receive(Equal(byte(1))))
		})
		Context("cfdevd stops after receiving uninstall command, thus closes connection before writing success code", func() {
			It("succeeds", func() {
				uninstallErrorCode = -1
				Expect(stopCmd.Execute()).To(Succeed())

				Eventually(instructions).Should(Receive(Equal(byte(1))))
			})
		})
		Context("cfdevd returns error to uninstall", func() {
			It("returns the error", func() {
				uninstallErrorCode = 1
				Expect(stopCmd.Execute()).To(MatchError("failed to uninstall cfdevd: errorcode: 1"))
			})
		})
	})
	Context("cfdevd socket is specified but does not exist", func() {
		It("succeeds", func() {
			Expect(stopCmd.Execute()).To(Succeed())
		})
	})
})
