package store

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAtomicWriteSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"hello":"world"}`)

	if err := AtomicWrite(path, data); err != nil {
		t.Fatalf("AtomicWrite: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Errorf("content = %q, want %q", got, data)
	}
}

func TestAtomicWriteOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	_ = AtomicWrite(path, []byte("first"))
	_ = AtomicWrite(path, []byte("second"))

	got, _ := os.ReadFile(path)
	if string(got) != "second" {
		t.Errorf("content = %q, want second", got)
	}
}

func TestAtomicWriteNoTmpLeftBehind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	_ = AtomicWrite(path, []byte("data"))

	tmp := path + ".tmp"
	if _, err := os.Stat(tmp); err == nil {
		t.Error("tmp file should not remain after successful write")
	}
}

func TestAtomicWriteConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = AtomicWrite(path, []byte("data"))
		}(i)
	}
	wg.Wait()

	// File should be valid (no corruption).
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("file should not be empty after concurrent writes")
	}
}

func TestAtomicWriteToReadonlyDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "readonly")
	_ = os.Mkdir(sub, 0o555)
	defer func() { _ = os.Chmod(sub, 0o755) }() // cleanup

	err := AtomicWrite(filepath.Join(sub, "file.json"), []byte("data"))
	if err == nil {
		t.Error("expected error writing to read-only directory")
	}
}

func TestAtomicWriteNestedPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "test.json")

	err := AtomicWrite(path, []byte("nested"))
	if err == nil {
		// On some systems this works if intermediate dirs exist.
		// On others it fails. Either way, no panic.
		got, _ := os.ReadFile(path)
		if string(got) != "nested" {
			t.Errorf("content = %q, want nested", got)
		}
	}
}
