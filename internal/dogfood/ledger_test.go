package dogfood

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLedgerAppendAndFindOpenByFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	rec := LedgerRecord{
		SchemaVersion: "dogfood-ledger.v1",
		EventID:       "evt-1",
		Fingerprint:   "fp-1",
		IssueURL:      "https://github.com/o/r/issues/1",
		Status:        "open",
		CreatedAt:     time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	}

	if err := l.Append(rec); err != nil {
		t.Fatal(err)
	}

	got, ok, err := l.FindOpenByFingerprint("fp-1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected open record, got ok=%v rec=%+v", ok, got)
	}
	if got.EventID != rec.EventID {
		t.Fatalf("expected event_id %q, got %q", rec.EventID, got.EventID)
	}
	if got.IssueURL == "" {
		t.Fatalf("expected issue_url to be set, got empty")
	}
}
