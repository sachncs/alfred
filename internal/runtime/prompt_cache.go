package runtime

import "sync"

// PromptCache provides a simple in-memory cache for compiled prompts.
// ponytail: just a sync.Map, swap for LRU if memory matters.
type PromptCache struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewPromptCache creates a cache.
func NewPromptCache() *PromptCache {
	return &PromptCache{m: make(map[string]string)}
}

// Get returns the cached prompt for a key, or empty string.
func (c *PromptCache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[key]
}

// Set stores a prompt under the given key.
func (c *PromptCache) Set(key, prompt string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = prompt
}

// Invalidate removes a cached entry.
func (c *PromptCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

// Clear removes all cached entries.
func (c *PromptCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m = make(map[string]string)
}
