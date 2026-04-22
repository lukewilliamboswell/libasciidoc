package docx_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDocx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docx Renderer Suite")
}
