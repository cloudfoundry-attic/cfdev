package env_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"os/exec"

	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
)

var _ = Describe("env", func() {
	Describe("BuildProxyConfig", func() {
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
				proxyConfig := env.BuildProxyConfig("bosh-ip", "router-ip", "host-ip")
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
				proxyConfig := env.BuildProxyConfig("bosh-ip", "router-ip", "host-ip")
				Expect(proxyConfig.Http).To(Equal("upper-some-http-proxy"))
				Expect(proxyConfig.Https).To(Equal("upper-some-https-proxy"))
				Expect(proxyConfig.NoProxy).To(Equal("upper-some-no-proxy,bosh-ip,router-ip,host-ip"))
			})
		})
	})

	Describe("CreateDirs", func() {
		var (
			dir         string
			logDir      string
			stateDir    string
			binaryDir   string
			cacheDir    string
			servicesDir string
			subject     env.Env
			conf        config.Config
		)

		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir(os.TempDir(), "test-space")
			Expect(err).NotTo(HaveOccurred())

			cacheDir = filepath.Join(dir, "cache")

			logDir = filepath.Join(dir, "log")
			os.MkdirAll(logDir, os.ModePerm)
			stateDir = filepath.Join(dir, "state")
			os.MkdirAll(stateDir, os.ModePerm)
			binaryDir = filepath.Join(dir, "bin")
			os.MkdirAll(binaryDir, os.ModePerm)
			servicesDir = filepath.Join(dir, "services")
			os.MkdirAll(servicesDir, os.ModePerm)

			conf = config.Config{
				LogDir:      logDir,
				StateDir:    stateDir,
				CacheDir:    cacheDir,
				BinaryDir:   binaryDir,
				ServicesDir: servicesDir,
			}

			subject = env.Env{
				Config: conf,
			}
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("re-creates the log and cache dir and remove the pre-existing others", func() {
			Expect(subject.CreateDirs()).To(Succeed())

			Expect(stateDir).NotTo(BeAnExistingFile())
			Expect(binaryDir).NotTo(BeAnExistingFile())
			Expect(binaryDir).NotTo(BeAnExistingFile())
			Expect(servicesDir).NotTo(BeAnExistingFile())

			Expect(logDir).To(BeAnExistingFile())
			Expect(cacheDir).To(BeAnExistingFile())
		})
	})

	Describe("SetupState", func() {
		var (
			tmpDir  string
			homeDir string
			subject env.Env
			conf    config.Config
		)

		BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir(os.TempDir(), "test-space-1")
			Expect(err).NotTo(HaveOccurred())

			homeDir, err = ioutil.TempDir(os.TempDir(), "test-space-2")
			Expect(err).NotTo(HaveOccurred())

			err = os.MkdirAll(filepath.Join(tmpDir, "dir"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tmpDir, "dir", "some-file"), []byte("some-content"), 0600)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command("tar", "cvzf", filepath.Join(tmpDir, "some-assetz.tgz"), "-C", tmpDir, "dir")
			output, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(output))

			conf = config.Config{
				CFDevHome: homeDir,
			}

			subject = env.Env{
				Config: conf,
			}
		})

		AfterEach(func() {
			os.RemoveAll(tmpDir)
			os.RemoveAll(homeDir)
		})

		It("unpacks everything in the cache dependency tarball", func() {
			Expect(subject.SetupState(filepath.Join(tmpDir, "some-assetz.tgz"))).To(Succeed())

			Expect(filepath.Join(homeDir, "dir")).To(BeADirectory())
			Expect(filepath.Join(homeDir, "dir", "some-file")).To(BeARegularFile())
		})
	})
})
