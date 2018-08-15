// +build windows

package hypervisor_test

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"code.cloudfoundry.org/cfdev/hypervisor"
	"math/rand"
	"time"
	"fmt"
)

var _ = Describe("HyperV", func() {
	var (
		hyperV     hypervisor.HyperV
		cfDevHome  string
		testIsoUrl = "https://s3.amazonaws.com/cfdev-test-assets/test.iso"
		err        error
		vmName string
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		vmName = randomVMName()
		
		cfDevHome, err = ioutil.TempDir("", "hypervtest")
		if err != nil {
			log.Fatal(err)
		}

		hyperV = hypervisor.HyperV{
			Config: config.Config{
				CFDevHome: cfDevHome,
				CacheDir:  filepath.Join(cfDevHome, "cache"),
			},
		}

		err = os.MkdirAll(hyperV.Config.CacheDir, 0666)
		Expect(err).ToNot(HaveOccurred())

		downloadAssets(hyperV.Config.CacheDir, testIsoUrl)
	})

	AfterEach(func() {
		err = os.RemoveAll(cfDevHome)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("CreateVM", func() {
		AfterEach(func() {
			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates hyperv VM", func() {
			vm := hypervisor.VM{
				Name: vmName,
				MemoryMB: 2000,
				CPUs:     1,
			}
			Expect(hyperV.CreateVM(vm)).To(Succeed())

			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property MemoryStartup,ProcessorCount",vmName))
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
				Expect(hyperV.Start(vmName)).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("State : Running"))

			})
			Context("when the vm is already running", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Start-VM -Name %s",vmName))
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())
					cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
					output, err := cmd.Output()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(ContainSubstring("State : Running"))
				})
				It("succeeds", func() {
					Expect(hyperV.Start(vmName)).To(Succeed())
				})
			})
		})
		Context("when the vm does not exist", func() {
			It("errors", func() {
				Expect(hyperV.Start(vmName)).To(MatchError(fmt.Sprintf("hyperv vm with name %s does not exist", vmName)))
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
					Expect(hyperV.Stop(vmName)).To(Succeed())
					cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-VM -Name %s | format-list -Property State", vmName))
					output, err := cmd.Output()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(ContainSubstring("State : Off"))
				})
			})

			Context("when the vm is not running", func() {
				It("succeeds", func() {
					Expect(hyperV.Stop(vmName)).To(Succeed())
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
				Expect(hyperV.Stop(vmName)).To(Succeed())
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
				Expect(hyperV.Destroy(vmName)).To(Succeed())
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
				Expect(hyperV.Destroy(vmName)).To(Succeed())
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
				Expect(hyperV.IsRunning(vmName)).To(BeFalse())
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
					Expect(hyperV.IsRunning(vmName)).To(BeFalse())
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
					Expect(hyperV.IsRunning(vmName)).To(BeTrue())
				})
			})
		})

	})
})

func downloadAssets(cacheDir string, isoSource string) {
	downloadFile(filepath.Join(cacheDir, "cfdev-efi.iso"), isoSource)
	downloadFile(filepath.Join(cacheDir, "cf-deps.iso"), isoSource)
}

func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomVMName() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return "some-vm" + string(b)
}
