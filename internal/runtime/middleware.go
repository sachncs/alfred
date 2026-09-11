package runtime

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// BearerAuth wraps an http.Handler and validates the Authorization header.
// If token is empty, auth is disabled (development mode).
func BearerAuth(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, `{"code":"unauthorized","message":"missing bearer token"}`, http.StatusUnauthorized)
			return
		}
		got := strings.TrimPrefix(auth, "Bearer ")
		// ponytail: constant-time comparison so the timing side-channel
		// cannot leak the configured token byte-by-byte. Length differences
		// are not exploitable for fixed-format bearer tokens.
		if len(got) != len(token) || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			http.Error(w, `{"code":"unauthorized","message":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
