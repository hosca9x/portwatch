package ports

import "sync"

// Tag represents a string label attached to a port entry.
type Tag string

// TagMap associates port entry keys with a set of tags.
type TagMap struct {
	mu   sync.RWMutex
	data map[string]map[Tag]struct{}
}

// NewTagMap creates an empty TagMap.
func NewTagMap() *TagMap {
	return &TagMap{
		data: make(map[string]map[Tag]struct{}),
	}
}

// Add attaches one or more tags to the given key.
func (t *TagMap) Add(key string, tags ...Tag) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.data[key]; !ok {
		t.data[key] = make(map[Tag]struct{})
	}
	for _, tag := range tags {
		t.data[key][tag] = struct{}{}
	}
}

// Remove detaches a tag from the given key.
func (t *TagMap) Remove(key string, tag Tag) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.data[key]; ok {
		delete(s, tag)
		if len(s) == 0 {
			delete(t.data, key)
		}
	}
}

// Tags returns a copy of the tag set for the given key.
func (t *TagMap) Tags(key string) []Tag {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.data[key]
	if !ok {
		return nil
	}
	out := make([]Tag, 0, len(s))
	for tag := range s {
		out = append(out, tag)
	}
	return out
}

// Has reports whether the key has the given tag.
func (t *TagMap) Has(key string, tag Tag) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if s, ok := t.data[key]; ok {
		_, found := s[tag]
		return found
	}
	return false
}

// Clear removes all tags for the given key.
func (t *TagMap) Clear(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.data, key)
}
