package site_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lukewilliamboswell/libasciidoc/internal/site"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustMkdirTemp(t *testing.T, prefix string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// loadTemplate
// ---------------------------------------------------------------------------

func TestLoadTemplate_MissingFile(t *testing.T) {
	err := site.Build(site.Config{
		SourceDir:    "/nonexistent-source-does-not-exist",
		OutputDir:    mustMkdirTemp(t, "out-*"),
		TemplatePath: "/nonexistent-template-does-not-exist.html",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadTemplate_BadSyntax(t *testing.T) {
	src := mustMkdirTemp(t, "src-*")
	out := mustMkdirTemp(t, "out-*")
	mustWriteFile(t, filepath.Join(src, "index.adoc"), "= Hello\n\nWorld.\n")

	tmplFile := filepath.Join(mustMkdirTemp(t, "tmpl-*"), "bad.html")
	mustWriteFile(t, tmplFile, `{{define "layout"}}{{.Unclosed{{end}}`)

	err := site.Build(site.Config{
		SourceDir:    src,
		OutputDir:    out,
		TemplatePath: tmplFile,
	})
	if err == nil {
		t.Fatal("expected error for bad template syntax, got nil")
	}
}

// ---------------------------------------------------------------------------
// Build — missing source directory
// ---------------------------------------------------------------------------

func TestBuild_MissingSourceDir(t *testing.T) {
	err := site.Build(site.Config{
		SourceDir: "/nonexistent-source-dir-xyz",
		OutputDir: mustMkdirTemp(t, "out-*"),
	})
	if err == nil {
		t.Fatal("expected error for missing source dir, got nil")
	}
}

// ---------------------------------------------------------------------------
// Build — renderPage error (source file that asciidoc cannot convert)
// We point the source dir at a file that exists but whose path we feed
// as a broken adoc reference by making a sub-dir unreadable after walk.
//
// Simpler approach: write a valid .adoc that references an include that
// doesn't exist, causing ConvertFile to return an error.
// ---------------------------------------------------------------------------

func TestBuild_RenderPageError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test permission errors as root")
	}
	if runtime.GOOS == "windows" {
		t.Skip("include error behaviour differs on Windows")
	}

	src := mustMkdirTemp(t, "src-*")
	out := mustMkdirTemp(t, "out-*")

	// Write a valid .adoc so walkSource succeeds and we have at least one page.
	mustWriteFile(t, filepath.Join(src, "index.adoc"), "= Good\n\nOK.\n")

	// Write a second .adoc that includes a non-existent file, which causes
	// ConvertFile to return an error (strict include processing).
	mustWriteFile(t, filepath.Join(src, "broken.adoc"), "= Broken\n\ninclude::does-not-exist.adoc[]\n")

	err := site.Build(site.Config{
		SourceDir: src,
		OutputDir: out,
	})
	if err == nil {
		t.Fatal("expected error from rendering broken.adoc, got nil")
	}
}

// ---------------------------------------------------------------------------
// walkSource — hidden directory is skipped (no error, just filtered out)
// ---------------------------------------------------------------------------

func TestWalkSource_HiddenDirSkipped(t *testing.T) {
	src := mustMkdirTemp(t, "src-*")
	out := mustMkdirTemp(t, "out-*")

	mustWriteFile(t, filepath.Join(src, "index.adoc"), "= Visible\n\nOK.\n")
	// A .adoc inside a hidden dir — should NOT be walked.
	mustWriteFile(t, filepath.Join(src, ".hidden", "secret.adoc"), "= Secret\n\nHide me.\n")

	err := site.Build(site.Config{
		SourceDir: src,
		OutputDir: out,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The hidden page should not appear in the output directory.
	if _, err := os.Stat(filepath.Join(out, ".hidden", "secret.html")); err == nil {
		t.Fatal("hidden directory content should not be copied to output")
	}
}

// ---------------------------------------------------------------------------
// copyFile — error: cannot create destination directory
// ---------------------------------------------------------------------------

func TestCopyFile_CannotCreateDestDir(t *testing.T) {
	if os.Getuid() == 0 || runtime.GOOS == "windows" {
		t.Skip("permission-based test not supported on this platform")
	}

	src := mustMkdirTemp(t, "src-*")
	out := mustMkdirTemp(t, "out-*")

	mustWriteFile(t, filepath.Join(src, "index.adoc"), "= Home\n\nOK.\n")
	// Write a static asset so copyAssets is called.
	mustWriteFile(t, filepath.Join(src, "asset.css"), "body{}")

	// Make the output dir read-only so MkdirAll inside copyFile fails.
	if err := os.Chmod(out, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(out, 0o755) //nolint:errcheck

	err := site.Build(site.Config{
		SourceDir: src,
		OutputDir: out,
	})
	if err == nil {
		t.Fatal("expected error when output dir is not writable, got nil")
	}
}

// ---------------------------------------------------------------------------
// copyFile — error: source file is unreadable (os.Open fails)
// ---------------------------------------------------------------------------

func TestCopyFile_SourceUnreadable(t *testing.T) {
	if os.Getuid() == 0 || runtime.GOOS == "windows" {
		t.Skip("permission-based test not supported on this platform")
	}

	src := mustMkdirTemp(t, "src-*")
	out := mustMkdirTemp(t, "out-*")

	mustWriteFile(t, filepath.Join(src, "index.adoc"), "= Home\n\nOK.\n")
	assetPath := filepath.Join(src, "style.css")
	mustWriteFile(t, assetPath, "body{}")

	// Make the asset unreadable so os.Open inside copyFile fails.
	if err := os.Chmod(assetPath, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(assetPath, 0o644) //nolint:errcheck

	err := site.Build(site.Config{
		SourceDir: src,
		OutputDir: out,
	})
	if err == nil {
		t.Fatal("expected error opening unreadable asset, got nil")
	}
}
