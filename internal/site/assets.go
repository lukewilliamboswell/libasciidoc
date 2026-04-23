package site

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// copyAssets copies all non-.adoc files from sourceDir to outputDir,
// preserving directory structure.
func copyAssets(sourceDir, outputDir string, assets []string) error {
	for _, relPath := range assets {
		src := filepath.Join(sourceDir, relPath)
		dst := filepath.Join(outputDir, relPath)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("copying asset %s: %w", relPath, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// walkSource walks the source directory and returns lists of .adoc file
// relative paths and asset (non-.adoc) relative paths.
func walkSource(sourceDir string) (adocFiles []string, assets []string, err error) {
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		// skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if strings.ToLower(filepath.Ext(path)) == ".adoc" {
			adocFiles = append(adocFiles, rel)
		} else {
			assets = append(assets, rel)
		}
		return nil
	})
	return
}
