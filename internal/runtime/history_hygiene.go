package runtime

import (
	"regexp"
	"strings"
)

// ansiRE matches ANSI escape sequences.
// ponytail: single regex, fast enough for prompt-size text.
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes all ANSI escape sequences from text.
func StripANSI(text string) string {
	return ansiRE.ReplaceAllString(text, "")
}

// HistoryHygiene cleans tool output and assistant text before storage.
type HistoryHygiene struct {
	stripANSI bool
	maxLen    int
}

// NewHistoryHygiene creates a hygiene processor.
// ponytail: defaults are sane, no config struct needed.
func NewHistoryHygiene() *HistoryHygiene {
	return &HistoryHygiene{stripANSI: true, maxLen: 100000}
}

// CleanText applies all hygiene passes to text.
func (h *HistoryHygiene) CleanText(text string) string {
	if h.stripANSI {
		text = StripANSI(text)
	}
	if h.maxLen > 0 && len(text) > h.maxLen {
		text = text[:h.maxLen] + "\n...[truncated]"
	}
	text = strings.TrimRight(text, "\n\r")
	return text
}

// HygieneMarker returns a marker string for compacted turns.
func HygieneMarker(turnsCompacted int) string {
	return "[history-hygiene: " + itoa(turnsCompacted) + " turns cleaned]"
}
