package ports

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsFirstCall(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	if !rl.Allow("tcp:8080") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestRateLimiter_BlocksWithinCooldown(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	rl.Allow("tcp:8080")
	if rl.Allow("tcp:8080") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestRateLimiter_AllowsAfterCooldown(t *testing.T) {
	rl := NewRateLimiter(10 * time.Millisecond)
	rl.Allow("tcp:9090")
	time.Sleep(20 * time.Millisecond)
	if !rl.Allow("tcp:9090") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	rl.Allow("tcp:8080")
	if !rl.Allow("udp:8080") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	rl.Allow("tcp:443")
	rl.Reset("tcp:443")
	if !rl.Allow("tcp:443") {
		t.Fatal("expected call after Reset to be allowed")
	}
}

func TestRateLimiter_Flush(t *testing.T) {
	rl := NewRateLimiter(5 * time.Second)
	rl.Allow("tcp:80")
	rl.Allow("tcp:443")
	rl.Flush()
	if !rl.Allow("tcp:80") {
		t.Fatal("expected call after Flush to be allowed for tcp:80")
	}
	if !rl.Allow("tcp:443") {
		t.Fatal("expected call after Flush to be allowed for tcp:443")
	}
}
