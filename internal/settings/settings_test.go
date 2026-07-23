package settings

import (
	"encoding/json"
	"testing"

	"github.com/alfred/alfred/internal/contract"
)

func TestDefaultAppSettings(t *testing.T) {
	s := DefaultAppSettings()
	if s.Version != "1" {
		t.Errorf("version: got %q", s.Version)
	}
	if s.Agents.ActiveRuntime != contract.AgentRuntimeAlfred {
		t.Errorf("active runtime: got %q", s.Agents.ActiveRuntime)
	}
	if s.LocalRuntime.Port != 8899 {
		t.Errorf("port: got %d", s.LocalRuntime.Port)
	}
	if s.AppBehavior.Theme != "dark" {
		t.Errorf("theme: got %q", s.AppBehavior.Theme)
	}
}

func TestNormalizeFillsDefaults(t *testing.T) {
	partial := contract.AppSettingsV1{
		Agents: contract.AgentSettings{
			ActiveRuntime: contract.AgentRuntimeOpencode,
		},
	}
	normalized := Normalize(partial)
	if normalized.Version != "1" {
		t.Errorf("version not filled: %q", normalized.Version)
	}
	if normalized.LocalRuntime.Port != 8899 {
		t.Errorf("port not filled: %d", normalized.LocalRuntime.Port)
	}
	if normalized.Agents.ActiveRuntime != contract.AgentRuntimeOpencode {
		t.Errorf("runtime overwritten: %q", normalized.Agents.ActiveRuntime)
	}
}

func TestMergeKeepsBase(t *testing.T) {
	base := DefaultAppSettings()
	base.AppBehavior.Theme = "light"
	base.LocalRuntime.Model = "claude-sonnet-4-20250514"

	partial := contract.AppSettingsV1{
		AppBehavior: contract.AppBehaviorSettings{
			Language: "fr",
		},
	}
	merged := Merge(base, partial)
	if merged.AppBehavior.Theme != "light" {
		t.Errorf("base theme lost: %q", merged.AppBehavior.Theme)
	}
	if merged.AppBehavior.Language != "fr" {
		t.Errorf("partial language not applied: %q", merged.AppBehavior.Language)
	}
	if merged.LocalRuntime.Model != "claude-sonnet-4-20250514" {
		t.Errorf("base model overwritten: %q", merged.LocalRuntime.Model)
	}
}

func TestStoreLoadSaveRoundTrip(t *testing.T) {
	store := newMockStore()

	// Load returns defaults when nothing saved
	s, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.AppBehavior.Theme != "dark" {
		t.Errorf("expected default theme, got %q", s.AppBehavior.Theme)
	}

	// Save
	s.AppBehavior.Theme = "light"
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	// Reload
	s2, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s2.AppBehavior.Theme != "light" {
		t.Errorf("expected saved theme, got %q", s2.AppBehavior.Theme)
	}
}

func TestStorePatch(t *testing.T) {
	store := newMockStore()

	partial := contract.AppSettingsV1{
		AppBehavior: contract.AppBehaviorSettings{
			Language: "de",
		},
	}
	merged, err := store.Patch(partial)
	if err != nil {
		t.Fatal(err)
	}
	if merged.AppBehavior.Language != "de" {
		t.Errorf("patch language: got %q", merged.AppBehavior.Language)
	}
	// Theme should still be default
	if merged.AppBehavior.Theme != "dark" {
		t.Errorf("patch lost theme: got %q", merged.AppBehavior.Theme)
	}
}

func TestNormalizeJSONRoundTrip(t *testing.T) {
	s := DefaultAppSettings()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var s2 contract.AppSettingsV1
	if err := json.Unmarshal(b, &s2); err != nil {
		t.Fatal(err)
	}
	s2 = Normalize(s2)
	if s2.LocalRuntime.Port != s.LocalRuntime.Port {
		t.Errorf("port mismatch after round-trip")
	}
}

// mockStore is an in-memory settings store for testing.
type mockStore struct {
	data map[string][]byte
}

func newMockStore() *SettingsStore {
	m := &mockStore{data: make(map[string][]byte)}
	return NewSettingsStore(m.get, m.set, m.del)
}

func (m *mockStore) get(namespace, key string) ([]byte, error) {
	v, ok := m.data[namespace+":"+key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockStore) set(namespace, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[namespace+":"+key] = b
	return nil
}

func (m *mockStore) del(namespace, key string) error {
	delete(m.data, namespace+":"+key)
	return nil
}
