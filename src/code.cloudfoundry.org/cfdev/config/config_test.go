package config_test

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	Describe("NewConfig", func() {
		Context("when CFDEV_HOME is not set", func() {
			var oldHome string

			BeforeEach(func() {
				oldHome = os.Getenv("HOME")
				os.Unsetenv("CFDEV_HOME")
				os.Setenv("HOME", "some-home-dir")
			})

			AfterEach(func() {
				os.Setenv("HOME", oldHome)
			})
			It("returns a config object with default values", func() {
				conf, err := config.NewConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(conf.BoshDirectorIP).To(Equal("10.245.0.2"))
				Expect(conf.CFRouterIP).To(Equal("10.144.0.34"))
				Expect(conf.CFDevHome).To(Equal(filepath.Join("some-home-dir", ".cfdev")))
				Expect(conf.StateDir).To(Equal(filepath.Join("some-home-dir", ".cfdev", "state")))
				Expect(conf.CacheDir).To(Equal(filepath.Join("some-home-dir", ".cfdev", "cache")))
				Expect(conf.LinuxkitPidFile).To(Equal(filepath.Join("some-home-dir", ".cfdev", "state", "linuxkit.pid")))
				Expect(conf.VpnkitPidFile).To(Equal(filepath.Join("some-home-dir", ".cfdev", "state", "vpnkit.pid")))
				Expect(conf.HyperkitPidFile).To(Equal(filepath.Join("some-home-dir", ".cfdev", "state", "hyperkit.pid")))
			})
		})
	})

	Context("when CFDEV_HOME is set", func() {
		BeforeEach(func() {
			os.Setenv("CFDEV_HOME", "some-cfdev-home")
		})

		AfterEach(func() {
			os.Unsetenv("CFDEV_HOME")
		})
		It("returns a config object with default values", func() {
			conf, err := config.NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(conf.BoshDirectorIP).To(Equal("10.245.0.2"))
			Expect(conf.CFRouterIP).To(Equal("10.144.0.34"))
			Expect(conf.CFDevHome).To(Equal(filepath.Join("some-cfdev-home")))
			Expect(conf.StateDir).To(Equal(filepath.Join("some-cfdev-home", "state")))
			Expect(conf.CacheDir).To(Equal(filepath.Join("some-cfdev-home", "cache")))
			Expect(conf.LinuxkitPidFile).To(Equal(filepath.Join("some-cfdev-home", "state", "linuxkit.pid")))
			Expect(conf.VpnkitPidFile).To(Equal(filepath.Join("some-cfdev-home", "state", "vpnkit.pid")))
			Expect(conf.HyperkitPidFile).To(Equal(filepath.Join("some-cfdev-home", "state", "hyperkit.pid")))
		})
	})
})
