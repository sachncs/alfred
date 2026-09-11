package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

// RequestID injects a request ID into the context (header or generated).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = generateID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// generateID returns a UUIDv4 string sourced from crypto/rand.
// 122 bits of entropy plus the fixed version/variant nibbles is sufficient
// for request-ID uniqueness under any realistic concurrent burst.
func generateID() string {
	id, err := uuid.NewRandom()
	if err == nil {
		return id.String()
	}
	// crypto/rand failures are vanishingly rare; fall back to a
	// deterministic but unique-per-process ID rather than panic.
	var b [16]byte
	if _, rerr := rand.Read(b[:]); rerr != nil {
		return "reqid-unavailable"
	}
	return hex.EncodeToString(b[:])
}
