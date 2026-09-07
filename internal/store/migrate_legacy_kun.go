package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// MigrateLegacyKun imports ~/.kun/config.json into the settings table.
//
// ponytail: best-effort, idempotent. If ~/.kun doesn't exist or settings
// already exist, returns nil. Records migration in `migrated_from_kun` key.
func MigrateLegacyKun(db *SQLite, home string) error {
	kunDir := filepath.Join(home, ".kun")
	configPath := filepath.Join(kunDir, "config.json")
	body, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if _, err := db.GetSetting("kun", "migrated_from_kun"); err == nil {
		return nil
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}

	for k, v := range raw {
		if err := db.SetSetting("kun", k, v); err != nil {
			return err
		}
	}
	return db.SetSetting("kun", "migrated_from_kun", true)
}
