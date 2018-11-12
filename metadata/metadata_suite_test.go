package metadata_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMetaData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MetaData Suite")
}
