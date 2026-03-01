package dogfood

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

func TestLedgerFindOpenByFingerprintIgnoresStaleOpenWhenLatestIsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	if err := l.Append(LedgerRecord{
		EventID:     "evt-open",
		Fingerprint: "fp-1",
		Status:      "open",
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	if err := l.Append(LedgerRecord{
		EventID:     "evt-closed",
		Fingerprint: "fp-1",
		Status:      "closed",
		CreatedAt:   time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	got, ok, err := l.FindOpenByFingerprint("fp-1")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected no open record after latest close, got %+v", got)
	}
}

func TestLedgerAppendConcurrentDoesNotLoseUpdates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	const writers = 100
	start := make(chan struct{})
	errs := make(chan error, writers)
	var wg sync.WaitGroup

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs <- l.Append(LedgerRecord{
				EventID:     fmt.Sprintf("evt-%03d", i),
				Fingerprint: "fp-concurrent",
				Status:      "open",
				CreatedAt:   time.Date(2026, 3, 1, 0, 0, i, 0, time.UTC),
			})
		}(i)
	}

	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("append error: %v", err)
		}
	}

	records, err := l.readAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != writers {
		t.Fatalf("expected %d records, got %d", writers, len(records))
	}
}

func TestLedgerFindOpenByFingerprintUsesQueueRetryIssueURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	if err := l.Append(LedgerRecord{
		EventID:     "evt-open",
		Fingerprint: "fp-queue",
		IssueURL:    "https://github.com/o/r/issues/1",
		Status:      LedgerStatusOpen,
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	queueURL := "https://github.com/o/r/issues/1"
	if err := l.Append(LedgerRecord{
		EventID:     "evt-retry",
		Fingerprint: "fp-queue",
		IssueURL:    queueURL,
		Status:      string(ActionQueueRetry),
		CreatedAt:   time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	got, ok, err := l.FindOpenByFingerprint("fp-queue")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected existing issue from queue_retry status")
	}
	if got.IssueURL != queueURL {
		t.Fatalf("expected issue_url %q, got %q", queueURL, got.IssueURL)
	}
}

func TestLedgerFindOpenByFingerprintDoesNotMaskOlderOpenWithRetryWithoutURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	openURL := "https://github.com/o/r/issues/10"
	if err := l.Append(LedgerRecord{
		EventID:     "evt-open",
		Fingerprint: "fp-retry",
		IssueURL:    openURL,
		Status:      LedgerStatusOpen,
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	if err := l.Append(LedgerRecord{
		EventID:     "evt-retry",
		Fingerprint: "fp-retry",
		Status:      string(ActionQueueRetry),
		CreatedAt:   time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	got, ok, err := l.FindOpenByFingerprint("fp-retry")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected older open issue to remain dedupe candidate")
	}
	if got.IssueURL != openURL {
		t.Fatalf("expected issue_url %q, got %q", openURL, got.IssueURL)
	}
}

func TestLedgerFindOpenByFingerprintRejectsOpenWithoutIssueURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	if err := l.Append(LedgerRecord{
		EventID:     "evt-bad-open",
		Fingerprint: "fp-bad-open",
		Status:      LedgerStatusOpen,
		CreatedAt:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	_, ok, err := l.FindOpenByFingerprint("fp-bad-open")
	if err == nil {
		t.Fatalf("expected error for open status without issue_url")
	}
	if ok {
		t.Fatalf("expected no dedupe candidate when open issue_url is malformed")
	}
}

func TestLedgerPersistsJSONArrayOfRecordsWithRequiredKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	createdAt := time.Date(2026, 3, 1, 2, 0, 0, 0, time.UTC)
	if err := l.Append(LedgerRecord{
		EventID:     "evt-shape",
		Fingerprint: "fp-shape",
		IssueURL:    "https://github.com/o/r/issues/99",
		Status:      LedgerStatusOpen,
		CreatedAt:   createdAt,
	}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted ledger: %v", err)
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode persisted ledger: %v", err)
	}

	records, ok := payload.([]any)
	if !ok {
		t.Fatalf("expected persisted ledger root to be json array, got %T", payload)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 persisted record, got %d", len(records))
	}

	record, ok := records[0].(map[string]any)
	if !ok {
		t.Fatalf("expected persisted record to be object, got %T", records[0])
	}

	for _, key := range []string{"schema_version", "event_id", "fingerprint", "status", "created_at"} {
		if _, ok := record[key]; !ok {
			t.Fatalf("expected persisted record key %q in %+v", key, record)
		}
	}

	if got, _ := record["schema_version"].(string); got != LedgerSchemaVersionV1 {
		t.Fatalf("expected schema_version %q, got %q", LedgerSchemaVersionV1, got)
	}
	if got, _ := record["status"].(string); got != LedgerStatusOpen {
		t.Fatalf("expected status %q, got %q", LedgerStatusOpen, got)
	}
}

func TestLedgerOmitsOptionalIssueURLWhenEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)

	if err := l.Append(LedgerRecord{
		EventID:     "evt-pending",
		Fingerprint: "fp-pending",
		Status:      string(ActionPendingReview),
		CreatedAt:   time.Date(2026, 3, 1, 2, 30, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted ledger: %v", err)
	}

	var records []map[string]any
	if err := json.Unmarshal(raw, &records); err != nil {
		t.Fatalf("decode persisted ledger: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 persisted record, got %d", len(records))
	}
	if _, ok := records[0]["issue_url"]; ok {
		t.Fatalf("did not expect issue_url key when empty, got %+v", records[0])
	}
}
