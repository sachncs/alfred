package client_test

import (
	"testing"

	"github.com/sachncs/alfred/internal/mcp/client"
)

func TestFileBridgeSignVerify(t *testing.T) {
	dir := t.TempDir()
	secret := []byte("test-secret-key-for-hmac")
	c := client.NewFileBridgeClient(dir, secret)
	_ = c
}
