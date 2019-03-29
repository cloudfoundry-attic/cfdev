package proxy_test

import (
	"fmt"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	proxyName string
)

var _ = Describe("cf dev proxy settings", func() {
	BeforeEach(func() {
		proxyName = os.Getenv("DOCKER_PROXY_NAME")
		if proxyName == "" {
			Skip("'DOCKER_PROXY_NAME' env var not set, skipping...")
		}
	})

	Context("when the HTTP_PROXY, HTTPS_PROXY, and NO_PROXY environment variables are set", func() {
		It("an app respect proxy environment variables", func() {
			Eventually(cf.Cf("login", "-a", "https://api.dev.cfdev.sh", "--skip-ssl-validation", "-u", "admin", "-p", "admin", "-o", "cfdev-org", "-s", "cfdev-space"), 5*time.Minute).Should(gexec.Exit(0))

			Eventually(func() int {
				sesh := cf.Cf("push", "cf-test-app", "-p", "../fixture", "-b", "ruby_buildpack")
				<-sesh.Exited
				return sesh.ExitCode()
			}, 10*time.Minute, 10*time.Second).Should(BeZero())

			By("making HTTP requests")
			Expect(httpGet("http://cf-test-app.dev.cfdev.sh/external")).To(ContainSubstring("Example Domain"))
			Eventually(fetchProxyLogs(proxyName)).Should(gbytes.Say(`Established connection to host ".*"`))

			By("making HTTPS requests")
			Expect(httpGet("http://cf-test-app.dev.cfdev.sh/external_https")).To(ContainSubstring("Example Domain"))
			Eventually(fetchProxyLogs(proxyName)).Should(gbytes.Say(`CONNECT .*:443 HTTP/1.1`))

			By("making a request from a site in the NO_PROXY list")
			Expect(httpGet("http://cf-test-app.dev.cfdev.sh/external_no_proxy")).To(ContainSubstring("www.google.com"))
			Consistently(fetchProxyLogs(proxyName)).ShouldNot(gbytes.Say(`Establish connection to host "google.com"`))
		})
	})
})

func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err
}

func fetchProxyLogs(proxyName string) *gbytes.Buffer {
	data, err := exec.Command("docker", "exec", "-t", proxyName, "tail", "-n", "10", "/logs/tinyproxy.log").Output()
	Expect(err).NotTo(HaveOccurred())
	return gbytes.BufferWithBytes(data)
}

func boshCurl(url string) {
	var command *exec.Cmd
	if runtime.GOOS != "windows" {
		command = exec.Command("bash", "-c", fmt.Sprintf(`eval "$(cf dev bosh env)" && bosh -d cf ssh api -c "curl %s"`, url))

	} else {
		command = exec.Command("powershell.exe", "-Command", fmt.Sprintf(`cf dev bosh env | Invoke-Expression; bosh -d cf ssh api -c "curl %s""`, url))
	}

	Expect(command.Run()).To(Succeed())

}
