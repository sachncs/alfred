package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Attachment is an uploaded attachment record.
type Attachment struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"threadId,omitempty"`
	MIMEType  string    `json:"mimeType"`
	Name      string    `json:"name,omitempty"`
	Size      int64     `json:"size"`
	Path      string    `json:"path,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// AttachmentStore stores uploaded attachments on disk.
type AttachmentStore struct {
	mu      sync.RWMutex
	dir     string
	counter int
	pending map[string]*Attachment
}

// NewAttachmentStore creates an attachment store rooted at dir.
func NewAttachmentStore() *AttachmentStore {
	return &AttachmentStore{pending: make(map[string]*Attachment)}
}

// SetDir sets the directory where attachments are persisted.
func (s *AttachmentStore) SetDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create attachment dir: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dir = dir
	return nil
}

// Save persists bytes as an attachment.
func (s *AttachmentStore) Save(threadID, mimeType, name string, body []byte) (*Attachment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dir == "" {
		return nil, fmt.Errorf("attachment store directory not configured")
	}
	s.counter++
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("rand: %w", err)
	}
	id := fmt.Sprintf("att-%d-%s", s.counter, hex.EncodeToString(idBytes))
	path := filepath.Join(s.dir, id)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return nil, fmt.Errorf("write attachment: %w", err)
	}
	att := &Attachment{
		ID:        id,
		ThreadID:  threadID,
		MIMEType:  mimeType,
		Name:      name,
		Size:      int64(len(body)),
		Path:      path,
		CreatedAt: time.Now().UTC(),
	}
	s.pending[id] = att
	return att, nil
}

// Get returns an attachment by ID.
func (s *AttachmentStore) Get(id string) (*Attachment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	att, ok := s.pending[id]
	return att, ok
}

// List returns all attachments.
func (s *AttachmentStore) List() []*Attachment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Attachment, 0, len(s.pending))
	for _, a := range s.pending {
		out = append(out, a)
	}
	return out
}
