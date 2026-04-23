package ports

import (
	"testing"
	"time"
)

func fixedEvictClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestEvictMap_SetAndGet(t *testing.T) {
	now := time.Now()
	em := newEvictMapWithClock(DefaultEvictMapPolicy(), fixedEvictClock(now))
	em.Set("k1", 42)
	v, ok := em.Get("k1")
	if !ok {
		t.Fatal("expected hit")
	}
	if v.(int) != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestEvictMap_MissingKeyReturnsFalse(t *testing.T) {
	em := NewEvictMap(DefaultEvictMapPolicy())
	_, ok := em.Get("missing")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestEvictMap_ExpiredEntryReturnsFalse(t *testing.T) {
	base := time.Now()
	policy := EvictMapPolicy{TTL: 1 * time.Second}
	current := base
	em := newEvictMapWithClock(policy, func() time.Time { return current })

	em.Set("k", "hello")
	current = base.Add(2 * time.Second) // advance past TTL

	_, ok := em.Get("k")
	if ok {
		t.Fatal("expected entry to be expired")
	}
}

func TestEvictMap_Delete(t *testing.T) {
	now := time.Now()
	em := newEvictMapWithClock(DefaultEvictMapPolicy(), fixedEvictClock(now))
	em.Set("k", "v")
	em.Delete("k")
	_, ok := em.Get("k")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestEvictMap_Purge(t *testing.T) {
	base := time.Now()
	policy := EvictMapPolicy{TTL: 10 * time.Second}
	current := base
	em := newEvictMapWithClock(policy, func() time.Time { return current })

	em.Set("a", 1)
	em.Set("b", 2)
	current = base.Add(20 * time.Second) // both expired
	em.Set("c", 3)                        // fresh entry

	removed := em.Purge()
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if em.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", em.Len())
	}
}

func TestEvictMap_LenCountsAll(t *testing.T) {
	now := time.Now()
	em := newEvictMapWithClock(DefaultEvictMapPolicy(), fixedEvictClock(now))
	em.Set("x", 1)
	em.Set("y", 2)
	if em.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", em.Len())
	}
}
