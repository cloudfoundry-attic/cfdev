package integration_test

import (
	"code.cloudfoundry.org/cfdev/servicew/client"
	"code.cloudfoundry.org/cfdev/servicew/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"
)

var _ = Describe("ServiceWrapper Lifecycle", func() {

	var (
		tempDir string
		swc     *client.ServiceWrapper
		label   = "org.cfdev.servicew.simple"
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cfdev-service-wrapper-")
		Expect(err).NotTo(HaveOccurred())

		swc = client.New(binaryPath, tempDir)
	})

	AfterEach(func() {
		swc.Stop(label)
		swc.Uninstall(label)
		os.RemoveAll(tempDir)
	})

	It("installs, runs, and remove services", func() {
		Expect(isRunning(swc, label)).To(BeFalse())

		contents, err := ioutil.ReadFile(fixturePath("simple.yml"))
		Expect(err).NotTo(HaveOccurred())

		var cfg config.Config
		yaml.Unmarshal(contents, &cfg)
		err = swc.Install(cfg)
		Expect(err).NotTo(HaveOccurred())

		Expect(isRunning(swc, label)).To(BeFalse())

		err = swc.Start(label)
		Expect(err).NotTo(HaveOccurred())

		Eventually(isRunning(swc, label), 10*time.Second).Should(BeTrue())

		if runtime.GOOS != "windows" {
			output := run("bash", "-c", "ps aux | grep 'sleep 12345'")
			Expect(strings.TrimSpace(output)).NotTo(BeEmpty())
		}

		err = swc.Stop(label)
		Expect(err).NotTo(HaveOccurred())
		Expect(isRunning(swc, label)).To(BeFalse())

		err = swc.Uninstall(label)
		Expect(isRunning(swc, label)).To(BeFalse())
	})
})

func isRunning(swc *client.ServiceWrapper, label string) bool {
	isRunning, err := swc.IsRunning(label)
	Expect(err).NotTo(HaveOccurred())
	return isRunning
}
