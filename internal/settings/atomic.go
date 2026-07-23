package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AtomicSettingsWrite writes AppSettingsV1 via tmp→rename for crash safety.
// ponytail: simple tmp→rename, survives concurrent writes at low frequency.
type AtomicSettingsWrite struct {
	dir string
}

// NewAtomicSettingsWriter creates an AtomicSettingsWrite.
func NewAtomicSettingsWriter(dir string) *AtomicSettingsWrite {
	return &AtomicSettingsWrite{dir: dir}
}

// Write persists AppSettingsV1 atomically.
func (a *AtomicSettingsWrite) Write(settings interface{}) error {
	if err := os.MkdirAll(a.dir, 0o755); err != nil {
		return fmt.Errorf("atomic settings: mkdir: %w", err)
	}

	b, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("atomic settings: marshal: %w", err)
	}

	p := filepath.Join(a.dir, "app_settings.json")
	tmp := p + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())

	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return fmt.Errorf("atomic settings: write tmp: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("atomic settings: rename: %w", err)
	}
	return nil
}

// Read reads AppSettingsV1 from disk.
func (a *AtomicSettingsWrite) Read(target interface{}) error {
	p := filepath.Join(a.dir, "app_settings.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}
