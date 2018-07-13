package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTcpbinder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfdevd Suite")
}
