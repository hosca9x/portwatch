package ports

import (
	"sync"
	"time"
)

// Debouncer delays forwarding of events until a quiet period has elapsed.
// If new events arrive before the timer fires, the timer is reset.
type Debouncer struct {
	mu      sync.Mutex
	delay   time.Duration
	clock   func() time.Time
	after   func(d time.Duration) <-chan time.Time
	timers  map[string]*time.Timer
	callback func(key string)
}

// NewDebouncer creates a Debouncer with the given quiet-period delay.
// callback is invoked once per key after the quiet period expires.
func NewDebouncer(delay time.Duration, callback func(key string)) *Debouncer {
	return &Debouncer{
		delay:    delay,
		timers:   make(map[string]*time.Timer),
		callback: callback,
		after:    time.NewTimer,
	}
}

// Trigger registers or resets the debounce timer for the given key.
func (d *Debouncer) Trigger(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
	}

	t := time.AfterFunc(d.delay, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
		d.callback(key)
	})
	d.timers[key] = t
}

// Pending returns true if a timer is currently active for the given key.
func (d *Debouncer) Pending(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.timers[key]
	return ok
}

// Cancel stops and removes any pending timer for the given key without
// invoking the callback.
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.timers[key]; ok {
		t.Stop()
		delete(d.timers, key)
	}
}
