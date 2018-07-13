package singlelinewriter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSinglelinewriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Singlelinewriter Suite")
}
