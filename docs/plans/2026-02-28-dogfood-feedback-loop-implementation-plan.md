# Dogfood Feedback Loop Skill Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a reusable cross-repo dogfooding feedback skill that captures failures, deduplicates noisy signals, and routes high-confidence feedback to GitHub issues.

**Architecture:** Implement a small `internal/dogfood` domain package with canonical event types, fingerprint/dedupe logic, router inference, ledger persistence, and GitHub publishing adapter. Expose it through `internal/tools/dogfoodfeedback` so any product repo can run it without touching the main `agentcli` command surface. Document operational usage in a new skill doc under `skills/` and wire schema checks for event/ledger contracts.

**Tech Stack:** Go 1.25.x, existing Taskfile CI gates, lightweight JSON schema checker (`internal/tools/schemacheck`), `gh` CLI integration via command adapter, markdown skill docs.

---

**Required supporting skills during execution:**
- `@superpowers:test-driven-development`
- `@superpowers:systematic-debugging`
- `@superpowers:verification-before-completion`

### Task 1: Create Canonical Dogfood Event Types

**Files:**
- Create: `internal/dogfood/types.go`
- Create: `internal/dogfood/types_test.go`

**Step 1: Write the failing test**

```go
func TestEventJSONIncludesRequiredKeys(t *testing.T) {
	e := Event{
		SchemaVersion: "dogfood-event.v1",
		EventID:       "evt-1",
		EventType:     EventTypeRuntimeError,
		SignalSource:  "local",
		Timestamp:     time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, key := range []string{"schema_version", "event_id", "event_type", "signal_source", "timestamp"} {
		if !strings.Contains(s, key) {
			t.Fatalf("missing key %s in %s", key, s)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestEventJSONIncludesRequiredKeys -v`
Expected: FAIL (package/files not found)

**Step 3: Write minimal implementation**

```go
type EventType string

const (
	EventTypeCIFailure   EventType = "ci_failure"
	EventTypeRuntimeError EventType = "runtime_error"
	EventTypeDocsDrift    EventType = "docs_drift"
)

type Event struct {
	SchemaVersion string    `json:"schema_version"`
	EventID       string    `json:"event_id"`
	EventType     EventType `json:"event_type"`
	SignalSource  string    `json:"signal_source"`
	Timestamp     time.Time `json:"timestamp"`
	RepoGuess     string    `json:"repo_guess,omitempty"`
	ErrorSummary  string    `json:"error_summary,omitempty"`
	EvidencePaths []string  `json:"evidence_paths,omitempty"`
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestEventJSONIncludesRequiredKeys -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/types.go internal/dogfood/types_test.go
git commit -m "feat(dogfood): add canonical event model"
```

### Task 2: Add Fingerprint + Dedupe Primitives

**Files:**
- Create: `internal/dogfood/fingerprint.go`
- Create: `internal/dogfood/fingerprint_test.go`
- Modify: `internal/dogfood/types.go`

**Step 1: Write the failing test**

```go
func TestFingerprintStableAcrossEvidenceOrder(t *testing.T) {
	a := Event{RepoGuess: "org/repo", EventType: EventTypeRuntimeError, SignalSource: "local", ErrorSummary: "panic: boom", EvidencePaths: []string{"b.log", "a.log"}}
	b := Event{RepoGuess: "org/repo", EventType: EventTypeRuntimeError, SignalSource: "local", ErrorSummary: "panic: boom", EvidencePaths: []string{"a.log", "b.log"}}
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatalf("fingerprint mismatch")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestFingerprintStableAcrossEvidenceOrder -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
func Fingerprint(e Event) string {
	evidence := append([]string(nil), e.EvidencePaths...)
	sort.Strings(evidence)
	base := strings.Join([]string{e.RepoGuess, string(e.EventType), e.SignalSource, normalizeError(e.ErrorSummary), strings.Join(evidence, ",")}, "|")
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:12])
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestFingerprintStableAcrossEvidenceOrder -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/types.go internal/dogfood/fingerprint.go internal/dogfood/fingerprint_test.go
git commit -m "feat(dogfood): add deterministic fingerprinting"
```

### Task 3: Implement Local Feedback Ledger

**Files:**
- Create: `internal/dogfood/ledger.go`
- Create: `internal/dogfood/ledger_test.go`

**Step 1: Write the failing test**

```go
func TestLedgerAppendAndFindOpenByFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dogfood-ledger.json")
	l := NewLedger(path)
	rec := LedgerRecord{EventID: "evt-1", Fingerprint: "fp-1", IssueURL: "https://github.com/o/r/issues/1", Status: "open"}
	if err := l.Append(rec); err != nil {
		t.Fatal(err)
	}
	got, ok, err := l.FindOpenByFingerprint("fp-1")
	if err != nil || !ok || got.IssueURL == "" {
		t.Fatalf("expected open record, got ok=%v err=%v rec=%+v", ok, err, got)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestLedgerAppendAndFindOpenByFingerprint -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
type LedgerRecord struct {
	SchemaVersion string    `json:"schema_version"`
	EventID       string    `json:"event_id"`
	Fingerprint   string    `json:"fingerprint"`
	IssueURL      string    `json:"issue_url,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Ledger struct { Path string }

func (l Ledger) Append(rec LedgerRecord) error { /* read+append+atomic write */ }
func (l Ledger) FindOpenByFingerprint(fp string) (LedgerRecord, bool, error) { /* linear scan */ }
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestLedgerAppendAndFindOpenByFingerprint -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/ledger.go internal/dogfood/ledger_test.go
git commit -m "feat(dogfood): add local ledger store and lookup"
```

### Task 4: Add Router Inference With Override and Confidence Gate

**Files:**
- Create: `internal/dogfood/router.go`
- Create: `internal/dogfood/router_test.go`
- Create: `internal/dogfood/router_config.go`

**Step 1: Write the failing test**

```go
func TestResolveRepoUsesOverrideFirst(t *testing.T) {
	r := Router{MinConfidence: 0.75}
	res := r.Resolve(RouteInput{OverrideRepo: "gh-xj/agentcli-go", CWD: "/tmp/x", GitRemote: "git@github.com:other/repo.git"})
	if res.Repo != "gh-xj/agentcli-go" || res.Confidence != 1.0 {
		t.Fatalf("unexpected route result: %+v", res)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestResolveRepoUsesOverrideFirst -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
type RouteInput struct {
	OverrideRepo string
	CWD          string
	GitRemote    string
}

type RouteResult struct {
	Repo       string
	Confidence float64
	Reason     string
	Pending    bool
}

func (r Router) Resolve(in RouteInput) RouteResult {
	if strings.TrimSpace(in.OverrideRepo) != "" {
		return RouteResult{Repo: in.OverrideRepo, Confidence: 1.0, Reason: "manual_override"}
	}
	// infer from git remote; if weak signal, set Pending=true
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestResolveRepoUsesOverrideFirst -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/router.go internal/dogfood/router_config.go internal/dogfood/router_test.go
git commit -m "feat(dogfood): add repo router with confidence gating"
```

### Task 5: Build GitHub Issue Publisher Adapter

**Files:**
- Create: `internal/dogfood/publisher.go`
- Create: `internal/dogfood/publisher_test.go`

**Step 1: Write the failing test**

```go
func TestPublisherCreateIssueWhenNoExistingOpenRecord(t *testing.T) {
	runner := &fakeRunner{out: "https://github.com/gh-xj/agentcli-go/issues/123\n"}
	pub := Publisher{Runner: runner}
	url, action, err := pub.Publish(PublishInput{Repo: "gh-xj/agentcli-go", Title: "dogfood: runtime error", Body: "details", ExistingIssueURL: ""})
	if err != nil || action != "created" || !strings.Contains(url, "/issues/123") {
		t.Fatalf("unexpected publish result action=%s url=%s err=%v", action, url, err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestPublisherCreateIssueWhenNoExistingOpenRecord -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
type CommandRunner interface {
	Run(name string, args ...string) (stdout string, err error)
}

type Publisher struct { Runner CommandRunner }

func (p Publisher) Publish(in PublishInput) (issueURL, action string, err error) {
	if in.ExistingIssueURL != "" {
		_, err := p.Runner.Run("gh", "issue", "comment", in.ExistingIssueURL, "--body", in.Body)
		return in.ExistingIssueURL, "commented", err
	}
	out, err := p.Runner.Run("gh", "issue", "create", "--repo", in.Repo, "--title", in.Title, "--body", in.Body)
	return strings.TrimSpace(out), "created", err
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestPublisherCreateIssueWhenNoExistingOpenRecord -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/publisher.go internal/dogfood/publisher_test.go
git commit -m "feat(dogfood): add github issue publisher adapter"
```

### Task 6: Add Decision Engine (Publish vs Pending-Review vs Queue)

**Files:**
- Create: `internal/dogfood/engine.go`
- Create: `internal/dogfood/engine_test.go`
- Modify: `internal/dogfood/types.go`

**Step 1: Write the failing test**

```go
func TestEngineMarksPendingWhenConfidenceBelowThreshold(t *testing.T) {
	eng := Engine{MinPublishConfidence: 0.8}
	dec := eng.Decide(DecisionInput{RepoConfidence: 0.4, Fingerprint: "fp-1"})
	if dec.Action != ActionPendingReview {
		t.Fatalf("expected pending-review, got %s", dec.Action)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/dogfood -run TestEngineMarksPendingWhenConfidenceBelowThreshold -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
type Action string

const (
	ActionCreateIssue   Action = "create_issue"
	ActionAppendComment Action = "append_comment"
	ActionPendingReview Action = "pending_review"
	ActionQueueRetry    Action = "queue_retry"
)

func (e Engine) Decide(in DecisionInput) Decision {
	if in.RepoConfidence < e.MinPublishConfidence {
		return Decision{Action: ActionPendingReview, Reason: "low_confidence_route"}
	}
	if in.HasOpenIssue {
		return Decision{Action: ActionAppendComment, Reason: "dedupe_open_issue"}
	}
	return Decision{Action: ActionCreateIssue, Reason: "new_fingerprint"}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/dogfood -run TestEngineMarksPendingWhenConfidenceBelowThreshold -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/dogfood/types.go internal/dogfood/engine.go internal/dogfood/engine_test.go
git commit -m "feat(dogfood): add publish decision engine"
```

### Task 7: Add `dogfoodfeedback` CLI Tool Entrypoint

**Files:**
- Create: `internal/tools/dogfoodfeedback/main.go`
- Create: `internal/tools/dogfoodfeedback/main_test.go`
- Modify: `Taskfile.yml`

**Step 1: Write the failing test**

```go
func TestRunRequiresEventFile(t *testing.T) {
	code := run([]string{"--repo-root", "."})
	if code != 2 {
		t.Fatalf("expected usage exit code 2, got %d", code)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tools/dogfoodfeedback -run TestRunRequiresEventFile -v`
Expected: FAIL

**Step 3: Write minimal implementation**

```go
func run(args []string) int {
	fs := flag.NewFlagSet("dogfoodfeedback", flag.ContinueOnError)
	eventPath := fs.String("event", "", "path to event json")
	ledgerPath := fs.String("ledger", ".docs/dogfood/ledger.json", "ledger file")
	overrideRepo := fs.String("repo", "", "override target repo (owner/name)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*eventPath) == "" {
		fmt.Fprintln(os.Stderr, "--event is required")
		return 2
	}
	// load event, route, decide, publish, append ledger
	return 0
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tools/dogfoodfeedback -run TestRunRequiresEventFile -v`
Expected: PASS

**Step 5: Add Taskfile helper commands**

```yaml
dogfood:dry-run:
  desc: Evaluate a dogfood event without publishing
  cmds:
    - go run ./internal/tools/dogfoodfeedback --event {{.EVENT}} --dry-run

dogfood:publish:
  desc: Publish a dogfood event to GitHub
  cmds:
    - go run ./internal/tools/dogfoodfeedback --event {{.EVENT}}
```

**Step 6: Run tool test and Taskfile linting checks**

Run: `go test ./internal/tools/dogfoodfeedback -v && task fmt:check`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/tools/dogfoodfeedback/main.go internal/tools/dogfoodfeedback/main_test.go Taskfile.yml
git commit -m "feat(dogfood): add dogfood feedback CLI tool"
```

### Task 8: Add Event/Ledger Schemas and Contract Fixtures

**Files:**
- Create: `schemas/dogfood-event.schema.json`
- Create: `schemas/dogfood-ledger.schema.json`
- Create: `testdata/contracts/dogfood-event.ok.json`
- Create: `testdata/contracts/dogfood-event.bad-missing-event-id.json`
- Create: `testdata/contracts/dogfood-ledger.ok.json`
- Create: `testdata/contracts/dogfood-ledger.bad-missing-status.json`
- Modify: `Taskfile.yml`

**Step 1: Write failing schema checks in Taskfile**

Add schema commands before creating files so checks fail.

**Step 2: Run gate to verify failure**

Run: `task schema:check`
Expected: FAIL for missing dogfood schema files

**Step 3: Write minimal schema + fixtures**

```json
{
  "schema_version": "dogfood-event.v1",
  "required_keys": ["schema_version", "event_id", "event_type", "signal_source", "timestamp", "fingerprint"]
}
```

```json
{
  "schema_version": "dogfood-ledger.v1",
  "required_keys": ["schema_version", "event_id", "fingerprint", "status", "created_at"]
}
```

**Step 4: Run schema checks to verify pass**

Run: `task schema:check && task schema:negative`
Expected: PASS

**Step 5: Commit**

```bash
git add schemas/dogfood-event.schema.json schemas/dogfood-ledger.schema.json testdata/contracts/dogfood-event.ok.json testdata/contracts/dogfood-event.bad-missing-event-id.json testdata/contracts/dogfood-ledger.ok.json testdata/contracts/dogfood-ledger.bad-missing-status.json Taskfile.yml
git commit -m "test(schema): add dogfood event and ledger contracts"
```

### Task 9: Publish Skill Docs and Repo Routing Guidance

**Files:**
- Create: `skills/dogfood-feedback-loop/SKILL.md`
- Create: `skills/dogfood-feedback-loop/examples/local-runtime-error.md`
- Create: `skills/dogfood-feedback-loop/examples/ci-failure.md`
- Create: `internal/tools/skillquality/dogfood_skill_test.go`
- Modify: `skill.md`
- Modify: `agents.md`

**Step 1: Write the failing docs check (manual assertion test)**

Create a small test to ensure required headings exist in the new skill doc.

```go
func TestDogfoodSkillHasRequiredSections(t *testing.T) {
	b, err := os.ReadFile("skills/dogfood-feedback-loop/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, h := range []string{"## In scope", "## Use this when", "## Feedback workflow", "## Replay"} {
		if !strings.Contains(s, h) {
			t.Fatalf("missing heading %s", h)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tools/skillquality -run TestDogfoodSkillHasRequiredSections -v`
Expected: FAIL (test/file missing)

**Step 3: Write minimal implementation**

```markdown
---
name: dogfood-feedback-loop
description: Capture dogfood failures and route high-confidence feedback to GitHub with dedupe and replay.
version: 1.0
---

## In scope
- capture CI/runtime/docs drift signals
- route to inferred repo with override support
- create issue or append comment by fingerprint
```

- Add two examples with event JSON and expected output.
- Register skill in `skill.md` index.
- Add operating notes in `agents.md` (not README).

**Step 4: Run docs checks**

Run: `task docs:check`
Expected: PASS

**Step 5: Commit**

```bash
git add skills/dogfood-feedback-loop/SKILL.md skills/dogfood-feedback-loop/examples/local-runtime-error.md skills/dogfood-feedback-loop/examples/ci-failure.md internal/tools/skillquality/dogfood_skill_test.go skill.md agents.md
git commit -m "docs(skill): add dogfood feedback loop skill and onboarding guidance"
```

### Task 10: Full Verification and Final Integration Commit

**Files:**
- Modify: `docs/plans/2026-02-28-dogfood-feedback-loop-implementation-plan.md` (checklist updates only)

**Step 1: Run full repository verification**

Run: `task ci`
Expected: PASS

**Step 2: Run local aggregate verification**

Run: `task verify`
Expected: PASS

**Step 3: Run focused dogfood package tests**

Run: `go test ./internal/dogfood ./internal/tools/dogfoodfeedback -v`
Expected: PASS

**Step 4: Update plan checklist with actual outcomes**

Record command outputs and any deviations directly in this plan file.

**Step 5: Commit final integration state**

```bash
git add docs/plans/2026-02-28-dogfood-feedback-loop-implementation-plan.md
git commit -m "chore(plan): mark dogfood feedback implementation verification results"
```

---

## Execution Results

### Implemented Commit Sequence

1. `fe7f2b9` - `feat(dogfood): add event model and deterministic fingerprinting`
2. `be67216` - `fix(dogfood): harden fingerprint canonicalization`
3. `c09f504` - `feat(dogfood): add ledger and repo router`
4. `c890a4c` - `fix(dogfood): harden ledger semantics and router parsing`
5. `55e7f0b` - `fix(dogfood): anchor github remote inference`
6. `487e9f8` - `feat(dogfood): add publisher engine and feedback tool`
7. `8d12e6c` - `fix(dogfood): harden publish dedupe and idempotency`
8. `d01eee2` - `fix(dogfood): lock idempotency marker updates`
9. `fec0ea8` - `fix(dogfood): recover stale idempotency locks`
10. `bcc029b` - `docs(dogfood): add schemas contracts and skill documentation`
11. `7216a23` - `fix(dogfood): align contracts and strengthen doc checks`

### Final Verification Outcomes

- `task ci` -> PASS
- `task verify` -> PASS
- `go test ./internal/dogfood ./internal/tools/dogfoodfeedback -v` -> PASS

### Notes

- Implemented via subagent-driven workflow with spec and code-quality review loops per task batch.
- Additional hardening commits were required for idempotency, lock recovery, and contract/doc consistency.
