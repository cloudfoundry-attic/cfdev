package env_test

import (
	"code.cloudfoundry.org/cfdev/resource"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

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

	Describe("CreateDirs", func() {
		var dir, homeDir, cacheDir, stateDir, boshDir, linuxkitDir, vpnkitStateDir, servicesDir, logDir string
		var err error
		var conf config.Config
		var subject env.Env

		BeforeEach(func() {
			dir, err = ioutil.TempDir(os.TempDir(), "test-space")
			Expect(err).NotTo(HaveOccurred())

			homeDir = filepath.Join(dir, "some-cfdev-home")
			cacheDir = filepath.Join(homeDir, "some-cache-dir")
			stateDir = filepath.Join(homeDir, "some-state-dir")
			boshDir = filepath.Join(stateDir, "some-bosh-state-dir")
			linuxkitDir = filepath.Join(stateDir, "some-linuxkit-state-dir")
			vpnkitStateDir = filepath.Join(stateDir, "some-vpnkit-state-dir")
			servicesDir = filepath.Join(homeDir, "services")
			logDir = filepath.Join(homeDir, "log")

			depsFile := filepath.Join(dir, "tmp-tar.tgz")
			conf = config.Config{
				CFDevHome:      homeDir,
				StateDir:       stateDir,
				StateBosh:      boshDir,
				StateLinuxkit:  linuxkitDir,
				CacheDir:       cacheDir,
				DepsFile:       &depsFile,
				VpnKitStateDir: vpnkitStateDir,
				ServicesDir:    servicesDir,
				LogDir:         logDir,
			}

			subject = env.Env{
				Config: conf,
			}
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("creates home, state, cache and services dirs", func() {
			Expect(subject.CreateDirs()).To(Succeed())
			_, err := os.Stat(homeDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(cacheDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(stateDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(linuxkitDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(vpnkitStateDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(servicesDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(logDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is already state in the home dir", func() {
			BeforeEach(func() {
				tmpDir, err := ioutil.TempDir(os.TempDir(), "tmp-tar")
				Expect(err).ToNot(HaveOccurred())

				Expect(os.Mkdir(homeDir, 0755)).To(Succeed())
				Expect(os.Mkdir(cacheDir, 0755)).To(Succeed())

				boshStateJson := filepath.Join(tmpDir, "state.json")
				Expect(ioutil.WriteFile(boshStateJson, []byte("state"), 0600)).To(Succeed())

				boshCreds := filepath.Join(tmpDir, "creds.yml")
				Expect(ioutil.WriteFile(boshCreds, []byte("creds"), 0600)).To(Succeed())

				boshSecret := filepath.Join(tmpDir, "secret")
				Expect(ioutil.WriteFile(boshSecret, []byte("some-bosh-secret"), 0600)).To(Succeed())

				boshJumpboxKey := filepath.Join(tmpDir, "jumpbox.key")
				Expect(ioutil.WriteFile(boshJumpboxKey, []byte("some-bosh-jumpbox-key"), 0600)).To(Succeed())

				boshCaCert := filepath.Join(tmpDir, "ca.crt")
				Expect(ioutil.WriteFile(boshCaCert, []byte("some-bosh-ca-cert"), 0600)).To(Succeed())

				tmpServicesDir := filepath.Join(tmpDir, "services")
				Expect(os.Mkdir(tmpServicesDir, 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpServicesDir, "service.file"), []byte("service file contents"), 0600)).To(Succeed())

				fpath := filepath.Join(tmpDir, "disk.qcow2")
				Expect(ioutil.WriteFile(fpath, []byte("tmp-disk"), 0600)).To(Succeed())

				tarDst, err := os.Create(*conf.DepsFile)
				Expect(err).ToNot(HaveOccurred())
				defer tarDst.Close()

				binaryPath := filepath.Join(tmpDir, "binaries")
				Expect(os.MkdirAll(binaryPath, 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(binaryPath, "binary.file"), []byte("one binary file"), 0600)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(binaryPath, "binary1.file"), []byte("other binary file"), 0600)).To(Succeed())

				deploymentConfigPath := filepath.Join(tmpDir, "deployment_config")
				Expect(os.MkdirAll(deploymentConfigPath, 0755)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(deploymentConfigPath, "director.yml"), []byte("some director config"), 0600)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(deploymentConfigPath, "cf.yml"), []byte("some cf configuration"), 0600)).To(Succeed())

				err = resource.Tar(tmpDir, tarDst)
				Expect(err).ToNot(HaveOccurred())
			})

			It("overwrites the qcow disk with a new one", func() {
				Expect(os.MkdirAll(filepath.Join(stateDir, "some-linuxkit-state-dir"), 0755)).To(Succeed())
				fpath := filepath.Join(stateDir, "some-linuxkit-state-dir", "disk.qcow2")
				Expect(ioutil.WriteFile(fpath, []byte("old-qcow"), 0600)).To(Succeed())

				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(stateDir, "some-linuxkit-state-dir", "disk.qcow2"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("tmp-disk"))
			})

			It("copies bosh state", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(stateDir, "some-bosh-state-dir", "state.json"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("state"))
			})

			It("copies bosh creds", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(stateDir, "some-bosh-state-dir", "creds.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("creds"))
			})

			It("copies services directory", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(servicesDir, "service.file"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("service file contents"))
			})

			It("copies binaries directory", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(cacheDir, "binary.file"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("one binary file"))

				b, err = ioutil.ReadFile(filepath.Join(cacheDir, "binary1.file"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("other binary file"))
			})

			It("copies deployment configuration directory", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(cacheDir, "director.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("some director config"))

				b, err = ioutil.ReadFile(filepath.Join(cacheDir, "cf.yml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("some cf configuration"))
			})

			It("copies bosh environment variables", func() {
				Expect(subject.CreateDirs()).To(Succeed())
				Expect(subject.SetupState()).To(Succeed())

				b, err := ioutil.ReadFile(filepath.Join(stateDir, "some-bosh-state-dir", "jumpbox.key"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("some-bosh-jumpbox-key"))

				b, err = ioutil.ReadFile(filepath.Join(stateDir, "some-bosh-state-dir", "ca.crt"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("some-bosh-ca-cert"))

				b, err = ioutil.ReadFile(filepath.Join(stateDir, "some-bosh-state-dir", "secret"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal("some-bosh-secret"))
			})
		})

		Context("when home dir cannot be created", func() {
			BeforeEach(func() {
				ioutil.WriteFile(homeDir, []byte{}, 0400)
			})

			It("returns an error", func() {
				err := subject.CreateDirs()
				Expect(err.Error()).To(ContainSubstring("failed to create dir"))
			})
		})
	})
})
