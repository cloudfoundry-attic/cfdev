package metadata_test

import (
	"code.cloudfoundry.org/cfdev/metadata"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetaData", func() {
	Context("reader returns", func() {
		var (
			metaDataPath string
		)

		BeforeEach(func() {
			tmp, err := ioutil.TempDir("", "tmp")
			Expect(err).ToNot(HaveOccurred())
			metaDataPath = filepath.Join(tmp, "metadata.yml")

			ioutil.WriteFile(metaDataPath, []byte(`---
compatibility_version: "v29"
default_memory: 8192
deployment_name: "cf"

splash_message: is simply dummy text

services:
- name: Mysql
  flag_name: mysql
  default_deploy: true
  handle: deploy-mysql
  script: bin/deploy-mysql
  deployment: cf-mysql

versions:
- name: some-release
  version: v123-some-version
- name: some-other-release
  version: v9.9.9`), 0777)
		})

		It("metadata", func() {
			metadata, err := metadata.New().Read(metaDataPath)

			Expect(err).ToNot(HaveOccurred())
			Expect(metadata.Version).To(Equal("v29"))
			Expect(metadata.Message).To(Equal("is simply dummy text"))
			Expect(metadata.Versions[0].Name).To(Equal("some-release"))
			Expect(metadata.Versions[0].Value).To(Equal("v123-some-version"))
		})
	})
})
