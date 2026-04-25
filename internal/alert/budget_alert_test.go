package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeBudgetReport(key string, used, max int) BudgetReport {
	return BudgetReport{
		Key:       key,
		Used:      used,
		Max:       max,
		Window:    time.Minute,
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestBudgetAlert_EmptyNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewBudgetAlerter(&buf)
	a.Notify(nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no output for nil reports")
	}
}

func TestBudgetAlert_SingleReportEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := NewBudgetAlerter(&buf)
	a.Notify([]BudgetReport{makeBudgetReport("host-a", 100, 100)})
	out := buf.String()
	if !strings.Contains(out, "[BUDGET]") {
		t.Fatalf("expected [BUDGET] prefix, got: %s", out)
	}
	if !strings.Contains(out, "host-a") {
		t.Fatalf("expected key in output, got: %s", out)
	}
}

func TestBudgetAlert_MultipleReportsMultipleLines(t *testing.T) {
	var buf bytes.Buffer
	a := NewBudgetAlerter(&buf)
	a.Notify([]BudgetReport{
		makeBudgetReport("x", 10, 10),
		makeBudgetReport("y", 5, 5),
	})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestBudgetAlert_LineContainsUsedAndMax(t *testing.T) {
	var buf bytes.Buffer
	a := NewBudgetAlerter(&buf)
	a.Notify([]BudgetReport{makeBudgetReport("z", 7, 20)})
	out := buf.String()
	if !strings.Contains(out, "used=7") {
		t.Fatalf("expected used=7 in output, got: %s", out)
	}
	if !strings.Contains(out, "max=20") {
		t.Fatalf("expected max=20 in output, got: %s", out)
	}
}

func TestNewBudgetAlerter_DefaultsToStdout(t *testing.T) {
	a := NewBudgetAlerter(nil)
	if a.out == nil {
		t.Fatal("expected non-nil writer when nil passed")
	}
}
