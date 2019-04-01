// +build windows

package hyperv_test

import (
	"code.cloudfoundry.org/cfdev/driver/hyperv"
	"code.cloudfoundry.org/cfdev/runner"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"math/rand"
	"time"
)

var _ = Describe("HyperV", func() {
	var (
		cfdevHome string
		hyperV    *hyperv.HyperV
		err       error
		vmName    string
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		vmName = randomVMName()

		cfdevHome, err = ioutil.TempDir("", "hypervtest")
		Expect(err).NotTo(HaveOccurred())

		hyperV = &hyperv.HyperV{
			Powershell: &runner.Powershell{},
			Config: config.Config{
				CFDevHome: cfdevHome,
				BinaryDir: filepath.Join(cfdevHome, "bin"),
				StateDir:  filepath.Join(cfdevHome, "state"),
			},
		}

		err = os.MkdirAll(hyperV.Config.BinaryDir, 0666)
		Expect(err).ToNot(HaveOccurred())
		err = os.MkdirAll(hyperV.Config.StateDir, 0666)
		Expect(err).ToNot(HaveOccurred())

		copyFile(
			filepath.Join(assetDir, "cfdev-efi-v2.iso"),
			filepath.Join(hyperV.Config.BinaryDir, "cfdev-efi-v2.iso"),
		)

		copyFile(
			filepath.Join(assetDir, "disk.vhdx"),
			filepath.Join(hyperV.Config.StateDir, "disk.vhdx"),
		)
	})

	AfterEach(func() {
		os.RemoveAll(cfdevHome)
	})

	Describe("CreateVM", func() {
		AfterEach(func() {
			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates hyperv VM", func() {
			Expect(hyperV.CreateVM(vmName, 1, 2000, filepath.Join(assetDir, "cfdev-efi-v2.iso"))).To(Succeed())

			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property MemoryStartup,ProcessorCount", vmName))
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10, 1).Should(gexec.Exit(0))
			Expect(session).To(gbytes.Say("MemoryStartup  : 2097152000"))
			Expect(session).To(gbytes.Say("ProcessorCount : 1"))

			cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VMHardDiskDrive -VMName %s", vmName))
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).ToNot(BeEmpty())
		})
	})

	Describe("Start", func() {
		Context("when the vm is already created", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			AfterEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Stop-VM -Name %s -Force", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())

				cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			It("starts the vm", func() {
				Expect(hyperV.StartVM(vmName)).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("State : Running"))

			})

			Context("when the vm is already running", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s", vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())

					cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
					output, err := cmd.Output()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(ContainSubstring("State : Running"))
				})

				It("succeeds", func() {
					Expect(hyperV.StartVM(vmName)).To(Succeed())
				})
			})
		})

		Context("when the vm does not exist", func() {
			It("errors", func() {
				Expect(hyperV.StartVM(vmName)).To(MatchError(fmt.Sprintf("hyperv vm with name %s does not exist", vmName)))
			})
		})
	})

	Describe("Stop", func() {
		Context("when the vm exists", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			AfterEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			Context("when the vm is running ", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s", vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())
				})

				It("stops the vm", func() {
					Expect(hyperV.StopVM(vmName)).To(Succeed())
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
					output, err := cmd.Output()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(ContainSubstring("State : Off"))
				})
			})

			Context("when the vm is not running", func() {
				It("succeeds", func() {
					Expect(hyperV.StopVM(vmName)).To(Succeed())
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s", vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit(0))
				})
			})
		})

		Context("when the vm does not exist", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s*", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
			It("succeeds", func() {
				Expect(hyperV.StopVM(vmName)).To(Succeed())
			})
		})
	})

	Describe("Destroy", func() {
		Context("when the vm exists and is stopped ", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			It("removes the vm", func() {
				Expect(hyperV.DestroyVM(vmName)).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s*", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
		})

		Context("when the vm does not exist", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s*", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
			It("succeeds", func() {
				Expect(hyperV.DestroyVM(vmName)).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s*", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
		})
	})

	Describe("IsRunning", func() {
		Context("when the vm does not exist", func() {
			It("returns false", func() {
				Expect(hyperV.IsVMRunning(vmName)).To(BeFalse())
			})
		})

		Context("when the vm exists", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			AfterEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			Context("when the vm exists and is not running", func() {
				It("returns false", func() {
					Expect(hyperV.IsVMRunning(vmName)).To(BeFalse())
				})
			})

			Context("when the vm is running", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s", vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())
				})

				AfterEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Stop-VM -Name %s -Force", vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())
				})

				It("returns true", func() {
					Expect(hyperV.IsVMRunning(vmName)).To(BeTrue())
				})
			})
		})
	})
})

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func copyFile(src, dest string) {
	srcFile, err := os.Open(src)
	Expect(err).NotTo(HaveOccurred())
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	Expect(err).NotTo(HaveOccurred())
	defer destFile.Close()

	io.Copy(destFile, srcFile)
}

func randomVMName() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return "some-vm" + string(b)
}
