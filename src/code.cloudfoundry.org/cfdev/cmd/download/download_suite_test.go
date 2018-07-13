package download_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDownload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Download Suite")
}
