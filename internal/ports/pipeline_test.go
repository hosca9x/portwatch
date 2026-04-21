package ports

import (
	"context"
	"testing"
	"time"
)

func sourceChan(entries []PortEntry) <-chan PortEntry {
	ch := make(chan PortEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func collectPipeline(ctx context.Context, ch <-chan PortEntry) []PortEntry {
	var out []PortEntry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func TestPipeline_NoStagesPassthrough(t *testing.T) {
	p := NewPipeline()
	entries := []PortEntry{{Port: 80, Proto: "tcp"}, {Port: 443, Proto: "tcp"}}
	ctx := context.Background()
	out := p.Run(ctx, sourceChan(entries))
	got := collectPipeline(ctx, out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestPipeline_FilterStageDropsEntries(t *testing.T) {
	p := NewPipeline()
	p.AddStage(FilterStage(func(e PortEntry) bool { return e.Port >= 1000 }))

	entries := []PortEntry{
		{Port: 80, Proto: "tcp"},
		{Port: 1080, Proto: "tcp"},
		{Port: 8080, Proto: "tcp"},
	}
	ctx := context.Background()
	out := p.Run(ctx, sourceChan(entries))
	got := collectPipeline(ctx, out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries after filter, got %d", len(got))
	}
	for _, e := range got {
		if e.Port < 1000 {
			t.Errorf("unexpected port %d passed filter", e.Port)
		}
	}
}

func TestPipeline_TransformStageModifiesEntries(t *testing.T) {
	p := NewPipeline()
	p.AddStage(TransformStage(func(e PortEntry) PortEntry {
		e.Proto = "udp"
		return e
	}))

	entries := []PortEntry{{Port: 53, Proto: "tcp"}}
	ctx := context.Background()
	out := p.Run(ctx, sourceChan(entries))
	got := collectPipeline(ctx, out)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Proto != "udp" {
		t.Errorf("expected proto udp, got %s", got[0].Proto)
	}
}

func TestPipeline_MultipleStagesChained(t *testing.T) {
	p := NewPipeline()
	// keep only port >= 100
	p.AddStage(FilterStage(func(e PortEntry) bool { return e.Port >= 100 }))
	// relabel proto
	p.AddStage(TransformStage(func(e PortEntry) PortEntry {
		e.Proto = "marked"
		return e
	}))

	entries := []PortEntry{
		{Port: 22, Proto: "tcp"},
		{Port: 443, Proto: "tcp"},
		{Port: 8080, Proto: "tcp"},
	}
	ctx := context.Background()
	out := p.Run(ctx, sourceChan(entries))
	got := collectPipeline(ctx, out)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	for _, e := range got {
		if e.Proto != "marked" {
			t.Errorf("expected proto marked, got %s", e.Proto)
		}
	}
}

func TestPipeline_CancelStopsProcessing(t *testing.T) {
	p := NewPipeline()
	p.AddStage(FilterStage(func(e PortEntry) bool { return true }))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// unbuffered infinite source would block forever without cancel support
	blocking := make(chan PortEntry)
	out := p.Run(ctx, blocking)

	done := make(chan struct{})
	go func() {
		for range out {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("pipeline did not stop after context cancel")
	}
}
