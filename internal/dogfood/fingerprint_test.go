package dogfood

import "testing"

func TestFingerprintStableAcrossEvidenceOrder(t *testing.T) {
	a := Event{
		RepoGuess:     "org/repo",
		EventType:     EventTypeRuntimeError,
		SignalSource:  "local",
		ErrorSummary:  "panic: boom",
		EvidencePaths: []string{"b.log", "a.log"},
	}
	b := Event{
		RepoGuess:     "org/repo",
		EventType:     EventTypeRuntimeError,
		SignalSource:  "local",
		ErrorSummary:  "panic: boom",
		EvidencePaths: []string{"a.log", "b.log"},
	}

	if Fingerprint(a) != Fingerprint(b) {
		t.Fatalf("fingerprint mismatch")
	}
}
