// +build windows

package process_test

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("HyperV process", func() {
	var (
		hyperV     process.HyperV
		cfDevHome  string
		testIsoUrl = "https://s3.amazonaws.com/cfdev-test-assets/test.iso"
		err        error
	)

	BeforeEach(func() {
		cfDevHome, err = ioutil.TempDir("", "hypervtest")
		if err != nil {
			log.Fatal(err)
		}

		hyperV = process.HyperV{
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
			cmd := exec.Command("powershell.exe", "-Command", "Remove-VM -Name cfdev -Force")
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates hyperv VM", func() {
			vm := process.VM{
				MemoryMB: 2000,
				CPUs:     1,
			}
			Expect(hyperV.CreateVM(vm)).To(Succeed())

			cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev | format-list -Property MemoryStartup,ProcessorCount")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10, 1).Should(gexec.Exit())
			Expect(session).To(gbytes.Say("MemoryStartup  : 2097152000"))
			Expect(session).To(gbytes.Say("ProcessorCount : 1"))

			cmd = exec.Command("powershell.exe", "-Command", "Get-VMHardDiskDrive -VMName cfdev")
			output, err := cmd.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).ToNot(BeEmpty())
		})
	})

	Describe("Stop", func() {
		Context("when the vm exists", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", "New-VM -Name cfdev -Generation 2 -NoVHD")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			AfterEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", "Remove-VM -Name cfdev -Force")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			Context("when the vm is running ", func() {
				BeforeEach(func() {
					cmd := exec.Command("powershell.exe", "-Command", "Start-VM -Name cfdev")
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit())
				})

				It("stops the vm", func() {
					Expect(hyperV.Stop("cfdev")).To(Succeed())
					cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev | format-list -Property State")
					output, err := cmd.Output()
					Expect(err).NotTo(HaveOccurred())
					Expect(string(output)).To(ContainSubstring("State : Off"))
				})
			})

			Context("when the vm is not running", func() {
				It("succeeds", func() {
					Expect(hyperV.Stop("cfdev")).To(Succeed())
					cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev")
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session, 10, 1).Should(gexec.Exit(0))
				})
			})
		})

		Context("when the vm does not exist", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev*")
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
			It("succeeds", func() {
				Expect(hyperV.Stop("cfdev")).To(Succeed())
			})
		})
	})

	Describe("Destroy", func() {
		Context("when the vm exists and is stopped ", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", "New-VM -Name cfdev -Generation 2 -NoVHD")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, 1).Should(gexec.Exit())
			})

			It("removes the vm", func() {
				Expect(hyperV.Destroy("cfdev")).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev*")
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
		})

		Context("when the vm does not exist", func() {
			BeforeEach(func() {
				cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev*")
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
			It("succeeds", func() {
				Expect(hyperV.Destroy("cfdev")).To(Succeed())
				cmd := exec.Command("powershell.exe", "-Command", "Get-VM -Name cfdev*")
				output, err := cmd.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
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
