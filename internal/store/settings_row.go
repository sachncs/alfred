package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// SettingsRow is one settings entry.
type SettingsRow struct {
	Namespace string
	Key       string
	Value     []byte
	UpdatedAt time.Time
}

// ErrSettingNotFound is returned when a setting is missing.
var ErrSettingNotFound = errors.New("setting not found")

// SetSetting upserts a setting.
func (s *SQLite) SetSetting(namespace, key string, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.Exec(`INSERT INTO settings (namespace, key, value_json, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(namespace, key) DO UPDATE SET value_json=excluded.value_json, updated_at=excluded.updated_at`,
		namespace, key, string(body), time.Now().UTC().Format(time.RFC3339))
	return err
}

// GetSetting returns a setting's raw JSON value.
func (s *SQLite) GetSetting(namespace, key string) ([]byte, error) {
	row := s.QueryRow(`SELECT value_json FROM settings WHERE namespace=? AND key=?`, namespace, key)
	var v string
	if err := row.Scan(&v); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSettingNotFound
		}
		return nil, err
	}
	return []byte(v), nil
}

// DeleteSetting removes a setting.
func (s *SQLite) DeleteSetting(namespace, key string) error {
	_, err := s.Exec(`DELETE FROM settings WHERE namespace=? AND key=?`, namespace, key)
	return err
}

// ListSettings returns all settings under a namespace.
func (s *SQLite) ListSettings(namespace string) ([]SettingsRow, error) {
	rows, err := s.Query(`SELECT namespace, key, value_json, updated_at FROM settings WHERE namespace=?`, namespace)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []SettingsRow
	for rows.Next() {
		var r SettingsRow
		var updated string
		if err := rows.Scan(&r.Namespace, &r.Key, &r.Value, &updated); err != nil {
			return nil, err
		}
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, r)
	}
	return out, rows.Err()
}
