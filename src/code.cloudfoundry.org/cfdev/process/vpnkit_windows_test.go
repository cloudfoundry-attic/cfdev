package process_test

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var _ = Describe("VpnKit", func() {
	var (
		vpnkit  *process.VpnKit
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cfdev-test-")
		Expect(err).NotTo(HaveOccurred())

		vpnkit = &process.VpnKit{
			Config: config.Config{
				CFDevHome: tempDir,
			},
		}
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Setup", func() {
		It("writes the dhcp and resolv conf files in the cfdevDir", func() {
			Expect(vpnkit.Setup()).To(Succeed())

			dnsPath := filepath.Join(tempDir, "resolv.conf")
			Expect(dnsPath).To(BeAnExistingFile())
			output, err := ioutil.ReadFile(dnsPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(MatchRegexp(`nameserver\s.*`))

			dhcpPath := filepath.Join(tempDir, "dhcp.json")
			Expect(dhcpPath).To(BeAnExistingFile())
			output, err = ioutil.ReadFile(dhcpPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("domainName"))
		})

		It("writes service guids to the registry", func() {
			Expect(vpnkit.Setup()).To(Succeed())

			command := exec.Command("powershell.exe", "-Command", `dir "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices"`)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			contents := string(session.Out.Contents())

			Expect(contents).To(ContainSubstring("CF Dev VPNkit Ethernet Service"))
			Expect(contents).To(ContainSubstring("CF Dev VPNkit Port Service"))
			Expect(contents).To(ContainSubstring("CF Dev VPNkit Forwarder Service"))
		})

	})

	FDescribe("VPNKIT", func() {
		var (
			vmGuid string
			vmName string
			vpnkit process.VpnKit
		)

		vmName = "testVm"
		vpnkit = process.VpnKit{}

		BeforeEach(func() {
			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("New-VM -Name %s -Generation 2 -NoVHD", vmName))
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())

			cmd = exec.Command("powershell.exe", "-Command", fmt.Sprintf("((Get-VM -Name %s).Id).Guid", vmName))
			output, err := cmd.Output()
			vmGuid = string(output)
		})

		AfterEach(func() {
			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Remove-VM -Name %s -Force", vmName))
			err := cmd.Run()
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates proper daemon spec", func() {
			Expect(vpnkit.Start()).To(Succeed())
		})
	})
})
