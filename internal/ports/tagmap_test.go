package ports

import (
	"sort"
	"testing"
)

func TestTagMap_AddAndHas(t *testing.T) {
	tm := NewTagMap()
	tm.Add("tcp:80", "web", "public")

	if !tm.Has("tcp:80", "web") {
		t.Error("expected tag 'web' to be present")
	}
	if !tm.Has("tcp:80", "public") {
		t.Error("expected tag 'public' to be present")
	}
	if tm.Has("tcp:80", "internal") {
		t.Error("unexpected tag 'internal'")
	}
}

func TestTagMap_Remove(t *testing.T) {
	tm := NewTagMap()
	tm.Add("tcp:443", "web", "tls")
	tm.Remove("tcp:443", "tls")

	if tm.Has("tcp:443", "tls") {
		t.Error("expected 'tls' to be removed")
	}
	if !tm.Has("tcp:443", "web") {
		t.Error("expected 'web' to remain")
	}
}

func TestTagMap_Tags(t *testing.T) {
	tm := NewTagMap()
	tm.Add("udp:53", "dns", "udp")

	tags := tm.Tags("udp:53")
	got := make([]string, len(tags))
	for i, tag := range tags {
		got[i] = string(tag)
	}
	sort.Strings(got)

	if len(got) != 2 || got[0] != "dns" || got[1] != "udp" {
		t.Errorf("unexpected tags: %v", got)
	}
}

func TestTagMap_TagsMissingKey(t *testing.T) {
	tm := NewTagMap()
	if tags := tm.Tags("tcp:9999"); tags != nil {
		t.Errorf("expected nil for missing key, got %v", tags)
	}
}

func TestTagMap_Clear(t *testing.T) {
	tm := NewTagMap()
	tm.Add("tcp:22", "ssh")
	tm.Clear("tcp:22")

	if tm.Has("tcp:22", "ssh") {
		t.Error("expected all tags cleared")
	}
	if tags := tm.Tags("tcp:22"); tags != nil {
		t.Errorf("expected nil after clear, got %v", tags)
	}
}

func TestTagMap_RemoveLastTagDeletesKey(t *testing.T) {
	tm := NewTagMap()
	tm.Add("tcp:8080", "proxy")
	tm.Remove("tcp:8080", "proxy")

	if tags := tm.Tags("tcp:8080"); tags != nil {
		t.Errorf("expected nil after removing last tag, got %v", tags)
	}
}
