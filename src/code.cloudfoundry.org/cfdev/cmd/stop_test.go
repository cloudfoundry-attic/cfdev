package cmd_test

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

type MockClient struct{}

func (mc *MockClient) Event(string, map[string]interface{}) error      { return nil }
func (mc *MockClient) Close()                                          {}
func (mc *MockClient) PromptOptIn(chan struct{}, cfanalytics.UI) error { return nil }

type MockLaunchdStop struct {
	stopLabels []string
	returns    map[string]error
}

func (m *MockLaunchdStop) Stop(label string) error {
	m.stopLabels = append(m.stopLabels, label)
	if v, ok := m.returns[label]; ok {
		return v
	}
	return nil
}

type MockCfdevdClient struct {
	uninstallWasCalled bool
	returns            error
}

func (m *MockCfdevdClient) Uninstall() (string, error) {
	m.uninstallWasCalled = true
	return "", m.returns
}

var _ = Describe("Stop", func() {
	var (
		cfg              config.Config
		stopCmd          *cobra.Command
		mockLaunchd      *MockLaunchdStop
		mockCfdevdClient *MockCfdevdClient
	)

	BeforeEach(func() {
		cfg = config.Config{
			Analytics: &MockClient{},
		}
		mockLaunchd = &MockLaunchdStop{
			returns: make(map[string]error, 0),
		}

		mockCfdevdClient = &MockCfdevdClient{}

		stopCmd = cmd.NewStop(cfg, mockLaunchd, mockCfdevdClient)
		stopCmd.SetArgs([]string{})
		stopCmd.SetOutput(GinkgoWriter)
	})

	It("stops linuxkt", func() {
		Expect(stopCmd.Execute()).To(Succeed())
		Expect(mockLaunchd.stopLabels).To(ContainElement(process.LinuxKitLabel))
	})

	It("stops vpnkit", func() {
		Expect(stopCmd.Execute()).To(Succeed())
		Expect(mockLaunchd.stopLabels).To(ContainElement(process.VpnKitLabel))
	})

	It("stops cfdevd", func() {
		Expect(stopCmd.Execute()).To(Succeed())
		Expect(mockCfdevdClient.uninstallWasCalled).To(BeTrue())
	})

	Context("stopping linuxkit fails", func() {
		BeforeEach(func() {
			mockLaunchd.returns[process.LinuxKitLabel] = fmt.Errorf("test")
		})
		It("stops the others and returns linuxkit error", func() {
			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop linuxkit: test"))

			Expect(mockLaunchd.stopLabels).To(ContainElement(process.LinuxKitLabel))
			Expect(mockLaunchd.stopLabels).To(ContainElement(process.VpnKitLabel))
			Expect(mockCfdevdClient.uninstallWasCalled).To(BeTrue())
		})
	})

	Context("stopping vpnkit fails", func() {
		BeforeEach(func() {
			mockLaunchd.returns[process.VpnKitLabel] = fmt.Errorf("test")
		})
		It("stops the others and returns vpnkit error", func() {
			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop vpnkit: test"))

			Expect(mockLaunchd.stopLabels).To(ContainElement(process.LinuxKitLabel))
			Expect(mockLaunchd.stopLabels).To(ContainElement(process.VpnKitLabel))
			Expect(mockCfdevdClient.uninstallWasCalled).To(BeTrue())
		})
	})

	Context("stopping cfdevd fails", func() {
		BeforeEach(func() {
			mockCfdevdClient.returns = fmt.Errorf("test")
		})
		It("stops the others and returns cfdevd error", func() {
			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to uninstall cfdevd: test"))

			Expect(mockLaunchd.stopLabels).To(ContainElement(process.LinuxKitLabel))
			Expect(mockLaunchd.stopLabels).To(ContainElement(process.VpnKitLabel))
			Expect(mockCfdevdClient.uninstallWasCalled).To(BeTrue())
		})
	})
})
