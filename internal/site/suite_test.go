package site_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Site Suite")
}
