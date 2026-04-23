package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- parseAttributes ---

func TestParseAttributes_Nil(t *testing.T) {
	got := parseAttributes(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %v", got)
	}
}

func TestParseAttributes_KeyOnly(t *testing.T) {
	got := parseAttributes([]string{"foo"})
	v, ok := got["foo"]
	if !ok || v != "" {
		t.Fatalf(`expected foo="", got %v`, got)
	}
}

func TestParseAttributes_KeyValue(t *testing.T) {
	got := parseAttributes([]string{"author=Jane"})
	if v, ok := got["author"]; !ok || v != "Jane" {
		t.Fatalf("expected author=Jane, got %v", got)
	}
}

func TestParseAttributes_ValueContainsEquals(t *testing.T) {
	// SplitN with n=2 means everything after the first '=' is the value.
	got := parseAttributes([]string{"url=http://example.com?a=b"})
	if v, ok := got["url"]; !ok || v != "http://example.com?a=b" {
		t.Fatalf("expected full value preserved, got %v", got)
	}
}

func TestParseAttributes_Multiple(t *testing.T) {
	got := parseAttributes([]string{"flag", "k=v", "x="})
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(got), got)
	}
	if got["flag"] != "" || got["k"] != "v" || got["x"] != "" {
		t.Fatalf("unexpected values: %v", got)
	}
}

// --- getOut helpers ---

func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	return cmd
}

// --- getOut ---

func TestGetOut_Stdout(t *testing.T) {
	cmd := newTestCmd()
	w, close := getOut(cmd, "file.adoc", "-", "html5")
	defer close() //nolint:errcheck
	if w == nil {
		t.Fatal("expected non-nil writer for stdout")
	}
}

func TestGetOut_NamedFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.html")
	cmd := newTestCmd()
	w, close := getOut(cmd, "file.adoc", outPath, "html5")
	if w == nil {
		t.Fatal("expected non-nil writer for named file")
	}
	if err := close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file to be created: %v", err)
	}
}

func TestGetOut_NamedFile_BadPath(t *testing.T) {
	cmd := newTestCmd()
	w, close := getOut(cmd, "file.adoc", "/nonexistent-dir-xyz-abc/out.html", "html5")
	defer close() //nolint:errcheck
	if w != nil {
		t.Fatal("expected nil writer for uncreateable output path")
	}
}

func TestGetOut_DerivedPath_HTML(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "doc.adoc")
	if err := os.WriteFile(srcPath, []byte("= Test"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := newTestCmd()
	w, close := getOut(cmd, srcPath, "", "html5")
	if w == nil {
		t.Fatal("expected non-nil writer for derived HTML path")
	}
	if err := close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	outPath := strings.TrimSuffix(srcPath, ".adoc") + ".html"
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected derived .html file to be created: %v", err)
	}
}

func TestGetOut_DerivedPath_DOCX(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "doc.adoc")
	if err := os.WriteFile(srcPath, []byte("= Test"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := newTestCmd()
	w, close := getOut(cmd, srcPath, "", "docx")
	if w == nil {
		t.Fatal("expected non-nil writer for derived DOCX path")
	}
	if err := close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	outPath := strings.TrimSuffix(srcPath, ".adoc") + ".docx"
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected derived .docx file to be created: %v", err)
	}
}

func TestGetOut_EmptySourcePath(t *testing.T) {
	// When sourcePath == "" and outputName == "", falls through to cmd.OutOrStdout().
	cmd := newTestCmd()
	w, close := getOut(cmd, "", "", "html5")
	defer close() //nolint:errcheck
	if w == nil {
		t.Fatal("expected non-nil writer (stdout fallback)")
	}
}

// --- NewVersionCmd ---

func TestNewVersionCmd_Runs(t *testing.T) {
	cmd := NewVersionCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	// Invoke Run directly — should not panic regardless of build info availability.
	cmd.Run(cmd, nil)
	// Produces at least one line of output.
	if buf.Len() == 0 {
		t.Fatal("expected some output from version command")
	}
}

// --- NewRootCmd ---

func TestNewRootCmd_NoArgs_PrintsHelp(t *testing.T) {
	root := NewRootCmd("ascii2html", "html5")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{})
	// No args → RunE calls cmd.Help() which returns nil.
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error with no args: %v", err)
	}
}

func TestNewRootCmd_MultipleInputs_NamedOutput_Error(t *testing.T) {
	root := NewRootCmd("ascii2html", "html5")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"-o", "out.html", "a.adoc", "b.adoc"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for multiple files with a named output file")
	}
}

func TestNewRootCmd_BadLogLevel_Error(t *testing.T) {
	root := NewRootCmd("ascii2html", "html5")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--log-level", "notavalidlevel", "somefile.adoc"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestNewRootCmd_StaticSite_MultipleArgs_Error(t *testing.T) {
	root := NewRootCmd("ascii2html", "html5")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--static-site", "dir1", "dir2"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for --static-site with multiple directory args")
	}
}

func TestNewRootCmd_DOCX_DefaultBackend(t *testing.T) {
	root := NewRootCmd("ascii2doc", "docx")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	// Just verify the command is constructed correctly (flags registered).
	if root.Use != "ascii2doc [flags] FILE" {
		t.Fatalf("unexpected Use: %q", root.Use)
	}
}
