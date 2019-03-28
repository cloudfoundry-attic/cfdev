package config_test

import (
	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("BuildProxyConfig", func() {
	var (
		cfg = config.Config{
			BoshDirectorIP: "bosh-ip",
			CFRouterIP:     "router-ip",
			HostIP:         "host-ip",
		}
	)

	Context("when proxy env vars are set", func() {
		BeforeEach(func() {
			os.Setenv("HTTP_PROXY", "some-http-proxy")
			os.Setenv("HTTPS_PROXY", "some-https-proxy")
			os.Setenv("NO_PROXY", "some-no-proxy")
		})

		AfterEach(func() {
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("NO_PROXY")
		})

		It("returns the http config", func() {
			proxyConfig := cfg.BuildProxyConfig()
			Expect(proxyConfig.Http).To(Equal("some-http-proxy"))
			Expect(proxyConfig.Https).To(Equal("some-https-proxy"))
			Expect(proxyConfig.NoProxy).To(Equal("some-no-proxy,bosh-ip,router-ip,host-ip"))
		})
	})

	Context("when multiple mixed case proxy envs prioritize uppercase", func() {
		BeforeEach(func() {
			os.Setenv("http_proxy", "lower-case-http-proxy")
			os.Setenv("HTTP_PROXY", "upper-some-http-proxy")
			os.Setenv("https_proxy", "lower-case-https-proxy")
			os.Setenv("HTTPS_PROXY", "upper-some-https-proxy")
			os.Setenv("no_proxy", "lower-some-no-proxy")
			os.Setenv("NO_PROXY", "upper-some-no-proxy,bosh-ip,router-ip")
		})

		AfterEach(func() {
			os.Unsetenv("http_proxy")
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("https_proxy")
			os.Unsetenv("HTTPS_PROXY")
			os.Unsetenv("no_proxy")
			os.Unsetenv("NO_PROXY")
		})

		It("returns the http config", func() {
			proxyConfig := cfg.BuildProxyConfig()
			Expect(proxyConfig.Http).To(Equal("upper-some-http-proxy"))
			Expect(proxyConfig.Https).To(Equal("upper-some-https-proxy"))
			Expect(proxyConfig.NoProxy).To(Equal("upper-some-no-proxy,bosh-ip,router-ip,host-ip"))
		})
	})
})
