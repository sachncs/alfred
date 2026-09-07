package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyKun(t *testing.T) {
	db := newTestSQLite(t)
	home := t.TempDir()

	// No ~/.kun — should be a no-op.
	if err := MigrateLegacyKun(db, home); err != nil {
		t.Fatal(err)
	}

	// Create ~/.kun/config.json.
	kunDir := filepath.Join(home, ".kun")
	_ = os.MkdirAll(kunDir, 0o755)
	body := []byte(`{"default_model":"gpt-4","theme":"dark"}`)
	if err := os.WriteFile(filepath.Join(kunDir, "config.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyKun(db, home); err != nil {
		t.Fatal(err)
	}

	v, err := db.GetSetting("kun", "default_model")
	if err != nil {
		t.Fatalf("default_model: %v", err)
	}
	if string(v) != `"gpt-4"` {
		t.Errorf("default_model = %q", v)
	}

	// Idempotent: run again, no duplicate.
	if err := MigrateLegacyKun(db, home); err != nil {
		t.Fatal(err)
	}
	rows, _ := db.ListSettings("kun")
	if len(rows) != 3 { // default_model, theme, migrated_from_kun
		t.Errorf("after re-run: rows = %d, want 3", len(rows))
	}
}
