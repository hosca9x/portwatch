package ports

import (
	"sync"
	"time"
)

// Suppressor prevents repeated alerts for the same port entry within a
// configurable quiet window. Once an entry is suppressed, it will not
// surface again until the window expires.
type Suppressor struct {
	mu      sync.Mutex
	window  time.Duration
	records map[string]time.Time
	now     func() time.Time
}

// NewSuppressor creates a Suppressor with the given quiet window duration.
func NewSuppressor(window time.Duration) *Suppressor {
	return &Suppressor{
		window:  window,
		records: make(map[string]time.Time),
		now:     time.Now,
	}
}

// IsSuppressed reports whether the entry identified by key has been seen
// within the quiet window. If not suppressed, the entry is recorded and
// true is returned on subsequent calls within the window.
func (s *Suppressor) IsSuppressed(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if last, ok := s.records[key]; ok {
		if s.now().Sub(last) < s.window {
			return true
		}
	}
	s.records[key] = s.now()
	return false
}

// Reset clears the suppression record for the given key, allowing the
// next occurrence to surface immediately.
func (s *Suppressor) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
}

// Flush removes all expired records, freeing memory for long-running
// daemons. It is safe to call concurrently.
func (s *Suppressor) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for k, t := range s.records {
		if now.Sub(t) >= s.window {
			delete(s.records, k)
		}
	}
}
