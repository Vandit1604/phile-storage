package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Vandit1604/phile-storage/internal/content"
)

// chdirTemp points the relative "data/" root at a throwaway directory.
func chdirTemp(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

func TestSaveAndReadBlockRoundTrip(t *testing.T) {
	chdirTemp(t)
	fs := NewFileStore("peer-a")

	data := []byte("content addressed bytes")
	c, err := content.Compute(data)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	if fs.HasBlock(c) {
		t.Fatal("block should not exist before save")
	}
	if err := fs.SaveBlock(c, bytes.NewReader(data)); err != nil {
		t.Fatalf("save: %v", err)
	}
	if !fs.HasBlock(c) {
		t.Fatal("block should exist after save")
	}

	got, err := fs.ReadBlock(c)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("round trip mismatch: got %q", got)
	}
}

// A CID-named path can never escape the peer's block directory, even if the
// original filename was hostile.
func TestBlockPathStaysInsidePeerDir(t *testing.T) {
	chdirTemp(t)
	fs := NewFileStore("peer-b")

	c, err := content.Compute([]byte("../../etc/passwd")) // hostile *content* still yields a safe CID
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if err := fs.SaveBlock(c, bytes.NewReader([]byte("x"))); err != nil {
		t.Fatalf("save: %v", err)
	}

	path, err := fs.GetBlockPath(c)
	if err != nil {
		t.Fatalf("get path: %v", err)
	}
	want := filepath.Join("data", "peer-b", "blocks")
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(want)) {
		t.Fatalf("block path %q escaped %q", path, want)
	}
}
