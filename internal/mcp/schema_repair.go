package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RepairSchema attempts to fix common JSON schema issues from model output.
// Returns the cleaned JSON and any validation errors found.
func RepairSchema(input string) (string, []string) {
	var errs []string
	s := strings.TrimSpace(input)

	// Strip markdown code fences
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		var cleaned []string
		for _, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "```") {
				continue
			}
			cleaned = append(cleaned, l)
		}
		s = strings.Join(cleaned, "\n")
		s = strings.TrimSpace(s)
		errs = append(errs, "stripped markdown code fences")
	}

	// Try to parse as-is
	if json.Valid([]byte(s)) {
		return s, errs
	}

	// Fix trailing commas: ,} → } and ,] → ]
	s = fixTrailingCommas(s)
	if json.Valid([]byte(s)) {
		errs = append(errs, "fixed trailing commas")
		return s, errs
	}

	// Fix single quotes → double quotes
	if strings.Contains(s, "'") {
		fixed := strings.ReplaceAll(s, "'", `"`)
		if json.Valid([]byte(fixed)) {
			errs = append(errs, "replaced single quotes with double quotes")
			return fixed, errs
		}
	}

	// Try to extract JSON from surrounding text
	if idx := strings.Index(s, "{"); idx >= 0 {
		end := strings.LastIndex(s, "}")
		if end > idx {
			candidate := s[idx : end+1]
			if json.Valid([]byte(candidate)) {
				errs = append(errs, fmt.Sprintf("extracted JSON from surrounding text (skipped %d chars)", idx))
				return candidate, errs
			}
		}
	}

	return s, append(errs, "could not repair schema")
}

func fixTrailingCommas(s string) string {
	var out []byte
	inString := false
	escape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escape {
			out = append(out, c)
			escape = false
			continue
		}
		if c == '\\' && inString {
			out = append(out, c)
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			out = append(out, c)
			continue
		}
		if inString {
			out = append(out, c)
			continue
		}
		if c == ',' {
			// Look ahead for } or ]
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\n' || s[j] == '\r' || s[j] == '\t') {
				j++
			}
			if j < len(s) && (s[j] == '}' || s[j] == ']') {
				continue // skip trailing comma
			}
		}
		out = append(out, c)
	}
	return string(out)
}
