package contract_test

import (
	"testing"

	"github.com/alfred/alfred/internal/contract"
)

func TestURLPolicyIsSafe(t *testing.T) {
	p := contract.NewURLPolicy()

	tests := []struct {
		url  string
		safe bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"mailto:user@example.com", true},
		{"ftp://example.com", false},
		{"javascript:alert(1)", false},
		{"://bad", false},
		{"", false},
		{"https://evil.com/steal", true},
	}
	for _, tt := range tests {
		if got := p.IsSafe(tt.url); got != tt.safe {
			t.Errorf("IsSafe(%q) = %v, want %v", tt.url, got, tt.safe)
		}
	}
}

func TestURLPolicyStrict(t *testing.T) {
	p := contract.NewURLPolicyStrict([]string{"github.com", "docs.python.org"})

	tests := []struct {
		url  string
		safe bool
	}{
		{"https://github.com/repo", true},
		{"https://docs.python.org/3/", true},
		{"http://github.com/repo", false},
		{"https://evil.com", false},
		{"https://github.com.evil.com/phish", false},
	}
	for _, tt := range tests {
		if got := p.IsSafe(tt.url); got != tt.safe {
			t.Errorf("IsSafe(%q) = %v, want %v", tt.url, got, tt.safe)
		}
	}
}
