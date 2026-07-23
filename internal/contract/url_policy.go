package contract

import "net/url"

// URLPolicy checks whether a URL is safe to render in the UI.
// ponytail: allowlist-based, covers schemes + hosts. extend when needed.
type URLPolicy struct {
	allowedSchemes map[string]bool
	allowedHosts   map[string]bool // nil = all hosts allowed for allowed schemes
}

// NewURLPolicy creates a policy allowing https and http.
func NewURLPolicy() *URLPolicy {
	return &URLPolicy{
		allowedSchemes: map[string]bool{"https": true, "http": true, "mailto": true},
	}
}

// NewURLPolicyStrict creates a policy allowing only https for specific hosts.
func NewURLPolicyStrict(hosts []string) *URLPolicy {
	p := NewURLPolicy()
	p.allowedSchemes = map[string]bool{"https": true}
	p.allowedHosts = make(map[string]bool, len(hosts))
	for _, h := range hosts {
		p.allowedHosts[h] = true
	}
	return p
}

// IsSafe reports whether rawURL passes the allowlist.
func (p *URLPolicy) IsSafe(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if !p.allowedSchemes[u.Scheme] {
		return false
	}
	if p.allowedHosts != nil && !p.allowedHosts[u.Hostname()] {
		return false
	}
	return true
}
