package ports

import (
	"testing"
	"time"
)

var fixedSamplerBase = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func makeSamplerEntries(ports ...int) []PortEntry {
	out := make([]PortEntry, len(ports))
	for i, p := range ports {
		out[i] = PortEntry{Port: p, Proto: "tcp"}
	}
	return out
}

func TestSampler_InitiallyEmpty(t *testing.T) {
	s := NewSampler(DefaultSamplePolicy())
	if s.Len() != 0 {
		t.Fatalf("expected 0 samples, got %d", s.Len())
	}
}

func TestSampler_RecordIncreasesLen(t *testing.T) {
	s := NewSampler(DefaultSamplePolicy())
	s.Record(makeSamplerEntries(80, 443))
	if s.Len() != 1 {
		t.Fatalf("expected 1 sample, got %d", s.Len())
	}
}

func TestSampler_ReturnsRecordedEntries(t *testing.T) {
	s := NewSampler(DefaultSamplePolicy())
	entries := makeSamplerEntries(22, 80)
	s.Record(entries)
	samples := s.Samples()
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample")
	}
	if len(samples[0].Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(samples[0].Entries))
	}
}

func TestSampler_EvictsOldSamples(t *testing.T) {
	now := fixedSamplerBase
	clock := func() time.Time { return now }
	policy := SamplePolicy{MaxSamples: 100, MaxAge: 5 * time.Minute}
	s := newSamplerWithClock(policy, clock)

	s.Record(makeSamplerEntries(80))
	now = now.Add(6 * time.Minute)
	s.Record(makeSamplerEntries(443))

	if s.Len() != 1 {
		t.Fatalf("expected old sample evicted, got %d samples", s.Len())
	}
	if s.Samples()[0].Entries[0].Port != 443 {
		t.Fatalf("expected remaining sample to be port 443")
	}
}

func TestSampler_CapsAtMaxSamples(t *testing.T) {
	now := fixedSamplerBase
	clock := func() time.Time { return now }
	policy := SamplePolicy{MaxSamples: 3, MaxAge: time.Hour}
	s := newSamplerWithClock(policy, clock)

	for i := 0; i < 10; i++ {
		s.Record(makeSamplerEntries(i))
		now = now.Add(time.Second)
	}

	if s.Len() != 3 {
		t.Fatalf("expected 3 samples (capped), got %d", s.Len())
	}
}

func TestSampler_SamplesReturnsCopy(t *testing.T) {
	s := NewSampler(DefaultSamplePolicy())
	s.Record(makeSamplerEntries(8080))
	a := s.Samples()
	a[0].Entries = nil
	b := s.Samples()
	if len(b[0].Entries) == 0 {
		t.Fatal("Samples() should return independent copy")
	}
}
