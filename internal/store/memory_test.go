package store

import (
	"testing"
)

func TestMemoryStoreSetAndGet(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	if err := s.Set(MemoryScopeUser, "theme", "dark"); err != nil {
		t.Fatal(err)
	}

	e, err := s.Get(MemoryScopeUser, "theme")
	if err != nil {
		t.Fatal(err)
	}
	if e.Value != "dark" {
		t.Errorf("value = %v, want dark", e.Value)
	}
}

func TestMemoryStoreGetNotFound(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	_, err := s.Get(MemoryScopeUser, "nope")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestMemoryStoreDelete(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	_ = s.Set(MemoryScopeUser, "key", "val")
	if err := s.Delete(MemoryScopeUser, "key"); err != nil {
		t.Fatal(err)
	}

	_, err := s.Get(MemoryScopeUser, "key")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestMemoryStoreList(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	_ = s.Set(MemoryScopeUser, "a", 1)
	_ = s.Set(MemoryScopeUser, "b", 2)
	_ = s.Set(MemoryScopeUser, "c", 3)

	entries, err := s.List(MemoryScopeUser)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("list: got %d entries, want 3", len(entries))
	}
}

func TestMemoryStoreScopeIsolation(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	_ = s.Set(MemoryScopeUser, "key", "user-val")
	_ = s.Set(MemoryScopeWorkspace, "key", "workspace-val")
	_ = s.Set(MemoryScopeProject, "key", "project-val")

	user, _ := s.Get(MemoryScopeUser, "key")
	ws, _ := s.Get(MemoryScopeWorkspace, "key")
	proj, _ := s.Get(MemoryScopeProject, "key")

	if user.Value != "user-val" {
		t.Errorf("user scope = %v, want user-val", user.Value)
	}
	if ws.Value != "workspace-val" {
		t.Errorf("workspace scope = %v, want workspace-val", ws.Value)
	}
	if proj.Value != "project-val" {
		t.Errorf("project scope = %v, want project-val", proj.Value)
	}
}

func TestMemoryStoreOverwrite(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	_ = s.Set(MemoryScopeUser, "key", "first")
	_ = s.Set(MemoryScopeUser, "key", "second")

	e, _ := s.Get(MemoryScopeUser, "key")
	if e.Value != "second" {
		t.Errorf("overwrite: got %v, want second", e.Value)
	}
}

func TestMemoryStoreComplexValue(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewMemoryStore(dir)

	val := map[string]any{"nested": true, "count": 42}
	_ = s.Set(MemoryScopeUser, "obj", val)

	e, _ := s.Get(MemoryScopeUser, "obj")
	if e.Value == nil {
		t.Fatal("expected non-nil value")
	}
}
