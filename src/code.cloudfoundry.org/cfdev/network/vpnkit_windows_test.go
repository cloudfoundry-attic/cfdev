package network_test

import (
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"code.cloudfoundry.org/cfdev/network"
	"os"
	"time"
)

var _ = Describe("VpnKit", func() {
	var (
		vpnkit  *network.VpnKit
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cfdev-test-")
		Expect(err).NotTo(HaveOccurred())

		vpnkit = &network.VpnKit{
			Config: config.Config{
				CFDevHome: tempDir,
			},
		}
	})

	AfterEach(func() {
		registryDeleteCmd := `Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" | ` +
			`Where-Object { $_.GetValue("ElementName") -match "CF Dev VPNKit" } | ` +
			`Foreach-Object { Remove-Item (Join-Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" $_.PSChildName) }`

		command := exec.Command("powershell.exe", "-Command", registryDeleteCmd)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 10 * time.Second).Should(gexec.Exit())
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
			Eventually(session, 10 * time.Second).Should(gexec.Exit())
			contents := string(session.Out.Contents())

			Expect(contents).To(ContainSubstring("CF Dev VPNkit Ethernet Service"))
			Expect(contents).To(ContainSubstring("CF Dev VPNkit Port Service"))
			Expect(contents).To(ContainSubstring("CF Dev VPNkit Forwarder Service"))
		})
	})
})
