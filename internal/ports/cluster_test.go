package ports

import (
	"testing"
	"time"
)

var fixedClusterClock = func() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

func makeClusterEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestCluster_BelowMinSizeNoReport(t *testing.T) {
	p := DefaultClusterPolicy()
	p.MinSize = 3
	p.MaxGap = 1
	ct := newClusterTrackerWithClock(p, fixedClusterClock)
	ct.Record(makeClusterEntry(80, "tcp"))
	ct.Record(makeClusterEntry(81, "tcp"))
	if got := ct.Clusters(); len(got) != 0 {
		t.Fatalf("expected no clusters, got %d", len(got))
	}
}

func TestCluster_AtMinSizeEmitsReport(t *testing.T) {
	p := DefaultClusterPolicy()
	p.MinSize = 3
	p.MaxGap = 1
	ct := newClusterTrackerWithClock(p, fixedClusterClock)
	for _, port := range []int{80, 81, 82} {
		ct.Record(makeClusterEntry(port, "tcp"))
	}
	reports := ct.Clusters()
	if len(reports) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(reports))
	}
	r := reports[0]
	if r.Start != 80 || r.End != 82 || r.Size != 3 {
		t.Errorf("unexpected cluster: %+v", r)
	}
	if r.Proto != "tcp" {
		t.Errorf("expected proto tcp, got %s", r.Proto)
	}
}

func TestCluster_GapSplitsClusters(t *testing.T) {
	p := DefaultClusterPolicy()
	p.MinSize = 2
	p.MaxGap = 1
	ct := newClusterTrackerWithClock(p, fixedClusterClock)
	for _, port := range []int{10, 11, 20, 21} {
		ct.Record(makeClusterEntry(port, "udp"))
	}
	reports := ct.Clusters()
	if len(reports) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(reports))
	}
}

func TestCluster_StaleEntriesEvicted(t *testing.T) {
	p := DefaultClusterPolicy()
	p.MinSize = 2
	p.MaxGap = 1
	p.TTL = 1 * time.Minute
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	ct := newClusterTrackerWithClock(p, clock)
	for _, port := range []int{100, 101, 102} {
		ct.Record(makeClusterEntry(port, "tcp"))
	}
	now = now.Add(2 * time.Minute)
	if got := ct.Clusters(); len(got) != 0 {
		t.Fatalf("expected no clusters after TTL, got %d", len(got))
	}
}

func TestCluster_MultipleProtosSeparate(t *testing.T) {
	p := DefaultClusterPolicy()
	p.MinSize = 2
	p.MaxGap = 1
	ct := newClusterTrackerWithClock(p, fixedClusterClock)
	for _, port := range []int{8080, 8081} {
		ct.Record(makeClusterEntry(port, "tcp"))
		ct.Record(makeClusterEntry(port, "udp"))
	}
	reports := ct.Clusters()
	if len(reports) != 2 {
		t.Fatalf("expected 2 clusters (one per proto), got %d", len(reports))
	}
}

func TestDefaultClusterPolicy_Values(t *testing.T) {
	p := DefaultClusterPolicy()
	if p.MaxGap <= 0 {
		t.Error("MaxGap should be positive")
	}
	if p.MinSize <= 0 {
		t.Error("MinSize should be positive")
	}
	if p.TTL <= 0 {
		t.Error("TTL should be positive")
	}
}
