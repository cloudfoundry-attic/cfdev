package env_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"fmt"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
)

var _ = Describe("env", func() {
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
			proxyConfig := env.BuildProxyConfig("bosh-ip", "router-ip")
			Expect(proxyConfig.Http).To(Equal("some-http-proxy"))
			Expect(proxyConfig.Https).To(Equal("some-https-proxy"))
			Expect(proxyConfig.NoProxy).To(Equal("some-no-proxy,bosh-ip,router-ip"))
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
			proxyConfig := env.BuildProxyConfig("bosh-ip", "router-ip")
			Expect(proxyConfig.Http).To(Equal("upper-some-http-proxy"))
			Expect(proxyConfig.Https).To(Equal("upper-some-https-proxy"))
			Expect(proxyConfig.NoProxy).To(Equal("upper-some-no-proxy,bosh-ip,router-ip"))
		})
	})

	Describe("SetupEnvironment", func() {
		var dir string
		var err error

		Context("Setup when the paths are writable", func() {
			BeforeEach(func() {
				dir, err = ioutil.TempDir(os.TempDir(), "test-space")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(dir)
			})

			It("Creates home, state, cache", func() {
				homeDir := filepath.Join(dir, "some-cfdev-home")
				cacheDir := filepath.Join(dir, "some-cache-dir")
				stateDir := filepath.Join(dir, "some-state-dir")

				conf := config.Config{
					CFDevHome: homeDir,
					CacheDir:  cacheDir,
					StateDir:  stateDir,
				}

				Expect(env.Setup(conf)).To(Succeed())
				_, err := os.Stat(homeDir)
				Expect(err).NotTo(HaveOccurred())
				_, err = os.Stat(cacheDir)
				Expect(err).NotTo(HaveOccurred())
				_, err = os.Stat(stateDir)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when setup fails", func() {
			var (
				dir      string
				homeDir  string
				cacheDir string
				stateDir string
			)
			BeforeEach(func() {
				dir, err = ioutil.TempDir(os.TempDir(), "test-space")
				Expect(err).NotTo(HaveOccurred())

				homeDir = filepath.Join(dir, "some-cfdev-hom")
				cacheDir = filepath.Join(dir, "some-cache-dir")
				stateDir = filepath.Join(dir, "some-state-dir")
			})

			AfterEach(func() {
				os.RemoveAll(dir)
			})

			Context("when home dir cannot be created", func() {
				BeforeEach(func() {
					ioutil.WriteFile(homeDir, []byte{}, 0400)
				})

				AfterEach(func() {
					os.RemoveAll(homeDir)
				})

				It("returns an error", func() {
					conf := config.Config{
						CFDevHome: homeDir,
						CacheDir:  cacheDir,
						StateDir:  stateDir,
					}

					err := env.Setup(conf)
					Expect(err.Error()).
						To(ContainSubstring(fmt.Sprintf("failed to create cfdevhome dir: path %s", homeDir)))
				})
			})

			Context("when cache dir cannot be created", func() {
				BeforeEach(func() {
					ioutil.WriteFile(cacheDir, []byte{}, 0400)
				})

				AfterEach(func() {
					os.RemoveAll(cacheDir)
				})

				It("returns an error", func() {
					conf := config.Config{
						CFDevHome: homeDir,
						CacheDir:  cacheDir,
						StateDir:  stateDir,
					}

					err := env.Setup(conf)
					Expect(err.Error()).
						To(ContainSubstring(fmt.Sprintf("failed to create cache dir: path %s", cacheDir)))
				})
			})

			Context("when state dir cannot be created", func() {
				BeforeEach(func() {
					ioutil.WriteFile(stateDir, []byte{}, 0400)
				})

				AfterEach(func() {
					os.RemoveAll(stateDir)
				})

				It("returns an error", func() {
					conf := config.Config{
						CFDevHome: homeDir,
						CacheDir:  cacheDir,
						StateDir:  stateDir,
					}

					err := env.Setup(conf)
					Expect(err.Error()).
						To(ContainSubstring(fmt.Sprintf("failed to create state dir: path %s", stateDir)))
				})
			})
		})
	})
})
