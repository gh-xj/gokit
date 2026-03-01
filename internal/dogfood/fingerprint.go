package dogfood

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func Fingerprint(e Event) string {
	evidence := append([]string(nil), e.EvidencePaths...)
	sort.Strings(evidence)

	base := strings.Join([]string{
		strings.TrimSpace(e.RepoGuess),
		string(e.EventType),
		strings.TrimSpace(e.SignalSource),
		normalizeErrorSummary(e.ErrorSummary),
		strings.Join(evidence, ","),
	}, "|")

	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:12])
}

func normalizeErrorSummary(summary string) string {
	return strings.ToLower(strings.Join(strings.Fields(summary), " "))
}
