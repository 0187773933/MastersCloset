// Package paths centralizes every real filesystem location v2 touches.
//
// v1 littered handlers and utils with string literals like "./v1/server/html/..."
// and "./logs", which only worked if the binary ran from a specific working
// directory (the build then cp -r'd those trees into Dropbox). v2 embeds all
// assets into the binary, so the only paths that remain are the data dirs
// resolved here, all under ~/.config/mct unless an absolute path is given.
package paths

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ConfigDir returns ~/.config/mct, creating it if needed.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "mct")
	os.MkdirAll(dir, 0755)
	return dir
}

// LogDir returns the directory daily log files are written to.
func LogDir() string {
	dir := filepath.Join(ConfigDir(), "logs")
	os.MkdirAll(dir, 0755)
	return dir
}

// Resolve returns p unchanged when absolute, otherwise anchors it under
// ConfigDir. Empty input returns empty so callers can detect "unset".
func Resolve(p string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(ConfigDir(), p)
}

// ResolveFingerprinted reproduces v1's per-machine data layout so an existing
// install is picked up automatically. For a relative configured name like
// "mct.db" it returns ~/.config/mct/mct_<fingerprint>.db; if that file/dir does
// not exist yet but the seed source does, it is copied there first (as v1's
// FixDBAndSearchIndex did). An absolute configured path is honored as-is (an
// explicit override, no fingerprinting).
func ResolveFingerprinted(configured, fingerprint string, isDir bool) string {
	if configured == "" {
		return ""
	}
	if filepath.IsAbs(configured) {
		return configured
	}
	base := filepath.Base(configured)
	ext := filepath.Ext(base)
	dst := filepath.Join(ConfigDir(), strings.TrimSuffix(base, ext)+"_"+fingerprint+ext)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		src, _ := filepath.Abs(configured)
		if src != dst {
			if _, err := os.Stat(src); err == nil {
				if isDir {
					CopyDir(src, dst)
				} else {
					CopyFile(src, dst)
				}
			}
		}
	}
	return dst
}

// CopyFile copies a single file, creating dst.
func CopyFile(src, dst string) error {
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
	_, err = io.Copy(out, in)
	return err
}

// CopyDir recursively copies a directory tree (used for the Bleve index).
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return CopyFile(path, target)
	})
}
