package proxy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestPrivileged(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cf dev - acceptance proxy - privileged suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5*time.Minute)
})