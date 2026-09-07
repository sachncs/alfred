package store

import "testing"

func TestSettingsCRUD(t *testing.T) {
	db := newTestSQLite(t)

	if err := db.SetSetting("ui", "theme", "dark"); err != nil {
		t.Fatal(err)
	}

	v, err := db.GetSetting("ui", "theme")
	if err != nil {
		t.Fatal(err)
	}
	if string(v) != `"dark"` {
		t.Errorf("value = %q", v)
	}

	// Overwrite.
	if err := db.SetSetting("ui", "theme", "light"); err != nil {
		t.Fatal(err)
	}
	v, _ = db.GetSetting("ui", "theme")
	if string(v) != `"light"` {
		t.Errorf("after overwrite = %q", v)
	}

	// List.
	rows, err := db.ListSettings("ui")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Errorf("list = %d", len(rows))
	}

	// Delete.
	if err := db.DeleteSetting("ui", "theme"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetSetting("ui", "theme"); err != ErrSettingNotFound {
		t.Errorf("after delete: err = %v", err)
	}
}

func TestSettingsNamespaces(t *testing.T) {
	db := newTestSQLite(t)

	_ = db.SetSetting("ui", "theme", "dark")
	_ = db.SetSetting("model", "default", "gpt-4")
	_ = db.SetSetting("ui", "lang", "en")

	uiRows, _ := db.ListSettings("ui")
	if len(uiRows) != 2 {
		t.Errorf("ui namespace = %d, want 2", len(uiRows))
	}
	modelRows, _ := db.ListSettings("model")
	if len(modelRows) != 1 {
		t.Errorf("model namespace = %d, want 1", len(modelRows))
	}
}
