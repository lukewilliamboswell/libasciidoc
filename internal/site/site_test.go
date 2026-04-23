package site_test

import (
	"os"
	"path/filepath"

	"github.com/lukewilliamboswell/libasciidoc/internal/site"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("static site builder", func() {

	var (
		sourceDir string
		outputDir string
	)

	BeforeEach(func() {
		var err error
		sourceDir, err = os.MkdirTemp("", "site-src-*")
		Expect(err).NotTo(HaveOccurred())
		outputDir, err = os.MkdirTemp("", "site-out-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(sourceDir)
		os.RemoveAll(outputDir)
	})

	writeFile := func(relPath, content string) {
		path := filepath.Join(sourceDir, relPath)
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		Expect(err).NotTo(HaveOccurred())
		err = os.WriteFile(path, []byte(content), 0o644)
		Expect(err).NotTo(HaveOccurred())
	}

	readOutput := func(relPath string) string {
		data, err := os.ReadFile(filepath.Join(outputDir, relPath))
		Expect(err).NotTo(HaveOccurred())
		return string(data)
	}

	Context("basic build", func() {

		It("should render a single page with default template", func() {
			writeFile("index.adoc", `= My Site

Welcome to my site.
`)
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).NotTo(HaveOccurred())

			content := readOutput("index.html")
			Expect(content).To(ContainSubstring("<title>My Site</title>"))
			Expect(content).To(ContainSubstring("Welcome to my site."))
		})

		It("should render nested directory structure", func() {
			writeFile("index.adoc", `= Home

Home page.
`)
			writeFile("guides/index.adoc", `= Guides
:weight: 1

Guide index.
`)
			writeFile("guides/getting-started.adoc", `= Getting Started
:weight: 1

Start here.
`)
			writeFile("guides/advanced.adoc", `= Advanced
:weight: 2

Advanced topics.
`)
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(readOutput("index.html")).To(ContainSubstring("<title>Home</title>"))
			Expect(readOutput("guides/index.html")).To(ContainSubstring("<title>Guides</title>"))
			Expect(readOutput("guides/getting-started.html")).To(ContainSubstring("Start here."))
			Expect(readOutput("guides/advanced.html")).To(ContainSubstring("Advanced topics."))
		})
	})

	Context("navigation", func() {

		It("should include navigation links in output", func() {
			writeFile("index.adoc", `= Home

Home page.
`)
			writeFile("about.adoc", `= About

About page.
`)
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).NotTo(HaveOccurred())

			content := readOutput("index.html")
			Expect(content).To(ContainSubstring(`<nav>`))
			Expect(content).To(ContainSubstring(`about.html`))
		})

		It("should order pages by weight then title", func() {
			writeFile("charlie.adoc", `= Charlie
:weight: 3

C.
`)
			writeFile("alpha.adoc", `= Alpha
:weight: 1

A.
`)
			writeFile("bravo.adoc", `= Bravo
:weight: 2

B.
`)
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).NotTo(HaveOccurred())

			content := readOutput("alpha.html")
			// Alpha should appear before Bravo which should appear before Charlie
			alphaIdx := indexOf(content, "Alpha")
			bravoIdx := indexOf(content, "Bravo")
			charlieIdx := indexOf(content, "Charlie")
			Expect(alphaIdx).To(BeNumerically("<", bravoIdx))
			Expect(bravoIdx).To(BeNumerically("<", charlieIdx))
		})
	})

	Context("static assets", func() {

		It("should copy non-adoc files to output", func() {
			writeFile("index.adoc", `= Home

Home.
`)
			writeFile("images/logo.png", "fake-png-data")
			writeFile("style.css", "body { color: red; }")

			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(readOutput("images/logo.png")).To(Equal("fake-png-data"))
			Expect(readOutput("style.css")).To(Equal("body { color: red; }"))
		})
	})

	Context("base path", func() {

		It("should prefix nav links with base path", func() {
			writeFile("index.adoc", `= Home

Home.
`)
			writeFile("about.adoc", `= About

About.
`)
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
				BasePath:  "/myrepo/",
			})
			Expect(err).NotTo(HaveOccurred())

			content := readOutput("index.html")
			Expect(content).To(ContainSubstring(`/myrepo/about.html`))
		})
	})

	Context("error handling", func() {

		It("should fail with no adoc files", func() {
			writeFile("readme.txt", "not an adoc file")
			err := site.Build(site.Config{
				SourceDir: sourceDir,
				OutputDir: outputDir,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no .adoc files"))
		})
	})
})

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
