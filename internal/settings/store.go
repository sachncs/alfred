package settings

import (
	"encoding/json"

	"github.com/alfred/alfred/internal/contract"
)

const settingsNamespace = "app"

// SettingsStore persists AppSettingsV1 via the SQLite settings table.
type SettingsStore struct {
	getter  func(namespace, key string) ([]byte, error)
	setter  func(namespace, key string, value any) error
	deleter func(namespace, key string) error
}

// NewSettingsStore creates a SettingsStore backed by SQLite settings table.
func NewSettingsStore(
	getter func(namespace, key string) ([]byte, error),
	setter func(namespace, key string, value any) error,
	deleter func(namespace, key string) error,
) *SettingsStore {
	return &SettingsStore{getter: getter, setter: setter, deleter: deleter}
}

// Load returns the persisted AppSettingsV1, or defaults if not found.
func (s *SettingsStore) Load() (contract.AppSettingsV1, error) {
	raw, err := s.getter(settingsNamespace, "app_settings")
	if err != nil {
		return DefaultAppSettings(), nil
	}
	var settings contract.AppSettingsV1
	if err := json.Unmarshal(raw, &settings); err != nil {
		return DefaultAppSettings(), nil
	}
	return Normalize(settings), nil
}

// Save persists the full AppSettingsV1.
func (s *SettingsStore) Save(settings contract.AppSettingsV1) error {
	return s.setter(settingsNamespace, "app_settings", settings)
}

// Patch merges a partial into the saved settings and persists.
func (s *SettingsStore) Patch(partial contract.AppSettingsV1) (contract.AppSettingsV1, error) {
	current, err := s.Load()
	if err != nil {
		return contract.AppSettingsV1{}, err
	}
	merged := Merge(current, partial)
	if err := s.Save(merged); err != nil {
		return contract.AppSettingsV1{}, err
	}
	return merged, nil
}

// Delete removes the saved settings (reverts to defaults on next Load).
func (s *SettingsStore) Delete() error {
	return s.deleter(settingsNamespace, "app_settings")
}
