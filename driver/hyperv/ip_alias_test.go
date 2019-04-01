package hyperv_test

import (
	"code.cloudfoundry.org/cfdev/driver/hyperv"
	"code.cloudfoundry.org/cfdev/runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"time"
)

var _ = Describe("IP Alias", func() {
	var (
		hyperV *hyperv.HyperV
	)

	BeforeEach(func() {
		hyperV = &hyperv.HyperV{
			Powershell: &runner.Powershell{},
		}
	})

	Describe("RemoveLoopbackAliases", func() {
		Context("when the switch exits", func() {
			BeforeEach(func() {
				command := exec.Command(
					"powershell.exe",
					"-Command",
					"New-VMSwitch -Name test-hostnet-vm-switch -SwitchType Internal -Notes 'Switch for CF Dev Networking'",
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, time.Second).Should(gexec.Exit(0))
				command = exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch*")
				output, err := command.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("test-hostnet-vm-switch"))
			})

			It("removes the switch", func() {
				Expect(hyperV.RemoveLoopbackAliases("test-hostnet-vm-switch")).To(Succeed())

				command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch*")
				output, err := command.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})
		})

		Context("when the switch does not exists", func() {
			BeforeEach(func() {
				command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch*")
				output, err := command.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})

			It("succeeds", func() {
				Expect(hyperV.RemoveLoopbackAliases("test-hostnet-vm-switch")).To(Succeed())
			})
		})
	})

	Describe("AddLoopbackAliases", func() {
		Context("when there is no switch", func() {
			BeforeEach(func() {
				command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch*")
				output, err := command.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(BeEmpty())
			})

			AfterEach(func() {
				command := exec.Command("powershell.exe", "-Command", "Remove-VMSwitch -Name test-hostnet-vm-switch -Force")
				Expect(command.Run()).To(Succeed())
			})

			It("we can bind and serve on these ips", func() {
				Expect(hyperV.AddLoopbackAliases("test-hostnet-vm-switch", "10.66.66.66", "10.22.33.44")).To(Succeed())
				testBindAndServe("10.66.66.66:6666", "10.22.33.44:5555")
			})

			It("is idempotent", func() {
				Expect(hyperV.AddLoopbackAliases("test-hostnet-vm-switch", "10.66.66.66", "10.22.33.44")).To(Succeed())
				Expect(hyperV.AddLoopbackAliases("test-hostnet-vm-switch", "10.66.66.66", "10.22.33.44")).To(Succeed())
				testBindAndServe("10.66.66.66:6666", "10.22.33.44:5555")
			})
		})

		Context("when the switch already exists", func() {
			BeforeEach(func() {
				command := exec.Command(
					"powershell.exe",
					"-Command",
					"New-VMSwitch -Name test-hostnet-vm-switch -SwitchType Internal -Notes 'Switch for CF Dev Networking'",
				)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, time.Second).Should(gexec.Exit(0))
				command = exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch*")
				output, err := command.Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(ContainSubstring("test-hostnet-vm-switch"))
			})

			AfterEach(func() {
				command := exec.Command("powershell.exe", "-Command", "Remove-VMSwitch -Name test-hostnet-vm-switch -Force")
				Expect(command.Run()).To(Succeed())
			})

			It("succeeds", func() {
				Expect(hyperV.AddLoopbackAliases("test-hostnet-vm-switch", "10.66.66.66", "10.22.33.44")).To(Succeed())

				command := exec.Command("powershell.exe", "-Command", "Get-VMSwitch test-hostnet-vm-switch | Measure-Object | Select -ExpandProperty Count")
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session, 10, time.Second).Should(gexec.Exit(0))
				Expect(session.Out).To(gbytes.Say("1"))

				testBindAndServe("10.66.66.66:6666", "10.22.33.44:5555")
			})
		})
	})
})

func testBindAndServe(addr string, addr1 string) {
	listener := serveString(addr, "some-response")
	defer listener.Close()

	listener1 := serveString(addr1, "some-other-response")
	defer listener1.Close()

	resp, err := http.Get("http://" + addr)
	Expect(err).NotTo(HaveOccurred())
	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(body)).To(ContainSubstring("some-response"))

	resp, err = http.Get("http://" + addr1)
	Expect(err).NotTo(HaveOccurred())
	body, err = ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(body)).To(ContainSubstring("some-other-response"))
}

func serveString(address, response string) net.Listener {
	listener, err := net.Listen("tcp4", address)
	Expect(err).NotTo(HaveOccurred())
	go func() {
		defer GinkgoRecover()
		http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(response))
		}))
	}()
	return listener
}
