package contract

// KeyboardShortcut is a single key binding.
type KeyboardShortcut struct {
	Command string `json:"command"`
	Binding string `json:"binding"`
}

// KeyboardShortcutCatalog holds all keyboard shortcuts.
type KeyboardShortcutCatalog struct {
	Shortcuts []KeyboardShortcut `json:"shortcuts"`
}

// HasConflict reports if two shortcuts share the same binding.
func (c *KeyboardShortcutCatalog) HasConflict(binding string) bool {
	n := 0
	for _, s := range c.Shortcuts {
		if s.Binding == binding {
			n++
		}
	}
	return n > 1
}
