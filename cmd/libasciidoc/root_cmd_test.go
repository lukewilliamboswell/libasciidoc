package main_test

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"

	main "github.com/lukewilliamboswell/libasciidoc/cmd/libasciidoc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("root cmd", func() {

	It("render with STDOUT output", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-o", "-", "test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())
	})

	It("render with file output", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		content, err := os.ReadFile("test/test.html")
		Expect(err).ToNot(HaveOccurred())
		Expect(content).ToNot(BeEmpty())
	})

	It("render docx with backend-aware file output", func() {
		// given
		dir := GinkgoT().TempDir()
		source := filepath.Join(dir, "sample.adoc")
		Expect(os.WriteFile(source, []byte("= Sample\n\ncontent"), 0o644)).To(Succeed())
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-b", "docx", source})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		content, err := os.ReadFile(filepath.Join(dir, "sample.docx"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content[:2])).To(Equal("PK"))
	})

	It("fail to parse bad log level", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"--log-level", "debug1", "-s", "test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).To(HaveOccurred())
	})

	It("render without header/footer", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-s", "-o", "-", "test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())
		Expect(buf.String()).ToNot(ContainSubstring(`<div id="footer">`))
	})

	It("render with attribute set", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-s", "-o", "-", "-afoo1=bar1", "-afoo2=bar2", "test/doc_with_attributes.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())
		Expect(buf.String()).To(Equal(`<div class="paragraph">
<p>bar1 and bar2</p>
</div>
`))
	})

	It("render with attribute reset", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-s", "-o", "-", "-afoo1=bar1", "-a!foo2", "test/doc_with_attributes.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).ToNot(BeEmpty())
		// console output also includes a warning message
		Expect(buf.String()).To(ContainSubstring(`unable to find entry for attribute with key 'foo2' in context"
<div class="paragraph">
<p>bar1 and {foo2}</p>
</div>
`))
	})

	It("render multiple files", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-s", "test/admonition.adoc", "test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("when rendering multiple files, return last error", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-s", "test/doesnotexist.adoc", "test/test.adoc"})
		// when
		err := root.Execute()
		// then
		Expect(err).To(HaveOccurred())
	})

	It("show help when executed with no arg", func() {
		// given
		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{})
		// when
		err := root.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("render docx with --theme flag and verify styled output", func() {
		dir := GinkgoT().TempDir()
		source := filepath.Join(dir, "themed.adoc")
		Expect(os.WriteFile(source, []byte("= Themed Doc\n\ncontent"), 0o644)).To(Succeed())
		theme := filepath.Join(dir, "theme.yml")
		Expect(os.WriteFile(theme, []byte("base:\n  font_family: Helvetica\n  font_size: 10.5\npage:\n  size: Letter\n"), 0o644)).To(Succeed())

		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-b", "docx", "--theme", theme, source})
		err := root.Execute()
		Expect(err).ToNot(HaveOccurred())

		// Read the output DOCX and verify themed values
		docxPath := filepath.Join(dir, "themed.docx")
		data, err := os.ReadFile(docxPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data[:2])).To(Equal("PK")) // valid zip

		// Open the zip and check styles.xml
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		Expect(err).ToNot(HaveOccurred())
		var stylesContent []byte
		var docContent []byte
		for _, f := range zr.File {
			rc, err := f.Open()
			Expect(err).ToNot(HaveOccurred())
			content, err := io.ReadAll(rc)
			Expect(err).ToNot(HaveOccurred())
			rc.Close()
			switch f.Name {
			case "word/styles.xml":
				stylesContent = content
			case "word/document.xml":
				docContent = content
			}
		}
		Expect(stylesContent).ToNot(BeEmpty())
		Expect(docContent).ToNot(BeEmpty())

		// Verify Helvetica font
		Expect(string(stylesContent)).To(ContainSubstring(`Helvetica`))
		// Verify 10.5pt = 21 half-points
		Expect(string(stylesContent)).To(ContainSubstring(`w:val="21"`))
		// Verify Letter page size (12240 x 15840)
		Expect(string(docContent)).To(ContainSubstring(`w:w="12240"`))
	})

	It("fail with invalid --theme path", func() {
		dir := GinkgoT().TempDir()
		source := filepath.Join(dir, "test.adoc")
		Expect(os.WriteFile(source, []byte("content"), 0o644)).To(Succeed())

		root := main.NewRootCmd()
		buf := new(bytes.Buffer)
		root.SetOutput(buf)
		root.SetArgs([]string{"-b", "docx", "--theme", "/nonexistent/theme.yml", "-o", "-", source})
		err := root.Execute()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no such file"))
	})

})
