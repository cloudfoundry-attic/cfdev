package acceptance

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"fmt"

	"path"

	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hyperkit starts and telemetry", func() {
	var (
		session *gexec.Session
		err     error
	)

	BeforeEach(func() {
		os.RemoveAll(filepath.Join(cfdevHome, "analytics"))
	})

	AfterEach(func() {
		session.Kill()
	})

	XIt("optout", func() {
		cmd := exec.Command(GetCfPluginPath(), "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "no")

		Eventually(func() ([]byte, error) {
			return ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))
		}).Should(MatchJSON(`{"enabled":false, "props":{"type":"cf"}}`))

	})

	XIt("optin", func() {
		cmd := exec.Command(GetCfPluginPath(), "dev", "start")
		inWriter, _ := cmd.StdinPipe()
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		fmt.Fprintln(inWriter, "yes")

		Eventually(func() ([]byte, error) {
			return ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))
		}).Should(MatchJSON(`{"enabled":true, "props":{"type":"cf"}}`))
	})

	XIt("is already opted in", func() {
		err := os.MkdirAll(path.Join(cfdevHome, "analytics"), 0755)
		Expect(err).ToNot(HaveOccurred())

		//TODO update the 'optin' parameter to reflect to new logic
		err = ioutil.WriteFile(path.Join(cfdevHome, "analytics", "analytics.txt"), []byte("optin"), 0755)
		Expect(err).ToNot(HaveOccurred())

		cmd := exec.Command(GetCfPluginPath(), "dev", "start")
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		//TODO wait 'till after deps.iso download to have test value
		Consistently(session, time.Second).ShouldNot(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		Expect(ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))).Should(MatchJSON([]byte(`{"enabled":true, "props":{"type":"cf"}}`)))
	})

	It("allows noninteractive telemetry --off command", func() {
		cmd := exec.Command(GetCfPluginPath(), "dev", "telemetry", "--off")
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Consistently(session, time.Second).ShouldNot(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		Expect(ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))).Should(MatchJSON([]byte(`{"cfAnalyticsEnabled": false,"customAnalyticsEnabled": false,"props":{}}`)))
	})

	It("allows noninteractive telemetry --on command", func() {
		cmd := exec.Command(GetCfPluginPath(), "dev", "telemetry", "--on")
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		Consistently(session, time.Second).ShouldNot(gbytes.Say("Are you ok with CF Dev periodically capturing anonymized telemetry"))

		Expect(ioutil.ReadFile(filepath.Join(cfdevHome, "analytics", "analytics.txt"))).Should(MatchJSON([]byte(`{"cfAnalyticsEnabled": true,"customAnalyticsEnabled": false,"props":{}}`)))
	})
})
