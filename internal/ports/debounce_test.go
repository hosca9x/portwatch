package ports

import (
	"sync"
	"testing"
	"time"
)

func TestDebouncer_CallbackFiredAfterDelay(t *testing.T) {
	var mu sync.Mutex
	fired := []string{}

	d := NewDebouncer(20*time.Millisecond, func(key string) {
		mu.Lock()
		fired = append(fired, key)
		mu.Unlock()
	})

	d.Trigger("port:8080")
	time.Sleep(40 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(fired) != 1 || fired[0] != "port:8080" {
		t.Fatalf("expected callback for port:8080, got %v", fired)
	}
}

func TestDebouncer_ResetOnRetrigger(t *testing.T) {
	var mu sync.Mutex
	count := 0

	d := NewDebouncer(30*time.Millisecond, func(_ string) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	d.Trigger("k")
	time.Sleep(15 * time.Millisecond)
	d.Trigger("k") // reset
	time.Sleep(15 * time.Millisecond)
	d.Trigger("k") // reset again
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Fatalf("expected callback once, got %d", count)
	}
}

func TestDebouncer_PendingReflectsState(t *testing.T) {
	d := NewDebouncer(50*time.Millisecond, func(_ string) {})

	if d.Pending("x") {
		t.Fatal("expected not pending before trigger")
	}
	d.Trigger("x")
	if !d.Pending("x") {
		t.Fatal("expected pending after trigger")
	}
	time.Sleep(80 * time.Millisecond)
	if d.Pending("x") {
		t.Fatal("expected not pending after delay elapsed")
	}
}

func TestDebouncer_CancelPreventsCallback(t *testing.T) {
	var mu sync.Mutex
	fired := false

	d := NewDebouncer(30*time.Millisecond, func(_ string) {
		mu.Lock()
		fired = true
		mu.Unlock()
	})

	d.Trigger("port:9090")
	d.Cancel("port:9090")
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if fired {
		t.Fatal("callback should not fire after cancel")
	}
}

func TestDebouncer_IndependentKeys(t *testing.T) {
	var mu sync.Mutex
	fired := map[string]int{}

	d := NewDebouncer(20*time.Millisecond, func(key string) {
		mu.Lock()
		fired[key]++
		mu.Unlock()
	})

	d.Trigger("a")
	d.Trigger("b")
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if fired["a"] != 1 || fired["b"] != 1 {
		t.Fatalf("expected one fire per key, got %v", fired)
	}
}
