package slotresource

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/dal"
	"github.com/gh-xj/agentops/resource"
)

// setupGitRepo creates a temporary git repo with an initial commit.
// Returns the repo path.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	// Create a subdirectory for the repo so sibling copies work
	repoDir := filepath.Join(tmp, "myrepo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("setup %v: %s: %v", args, out, err)
		}
	}
	return repoDir
}

// newTestResource creates a SlotResource wired to real dal implementations
// with the projectDir override set.
func newTestResource(t *testing.T, projectDir string) (*SlotResource, *agentops.AppContext) {
	t.Helper()
	sr := New(&realFS{}, &realExec{})
	ctx := agentops.NewAppContext(context.Background())
	ctx.Values["project_dir"] = projectDir
	return sr, ctx
}

// realFS implements dal.FileSystem using the real OS.
type realFS struct{}

func (f *realFS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (f *realFS) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func (f *realFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f *realFS) WriteFile(path string, data []byte, perm int) error {
	return os.WriteFile(path, data, os.FileMode(perm))
}

func (f *realFS) ReadDir(path string) ([]dal.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var result []dal.DirEntry
	for _, e := range entries {
		result = append(result, dal.DirEntry{Name: e.Name(), IsDir: e.IsDir()})
	}
	return result, nil
}

func (f *realFS) BaseName(path string) string {
	return filepath.Base(path)
}

// realExec implements dal.Executor using real os/exec.
type realExec struct{}

func (e *realExec) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (e *realExec) RunInDir(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (e *realExec) RunOsascript(script string) string { return "" }
func (e *realExec) Which(cmd string) bool             { return false }

// --- Interface compliance ---

func TestInterfaceCompliance(t *testing.T) {
	var _ resource.Resource = (*SlotResource)(nil)
	var _ resource.Deleter = (*SlotResource)(nil)
	var _ resource.Syncer = (*SlotResource)(nil)
	var _ resource.Doctor = (*SlotResource)(nil)
	var _ resource.Pruner = (*SlotResource)(nil)
}

// --- Schema ---

func TestSchema(t *testing.T) {
	sr := New(&realFS{}, &realExec{})
	s := sr.Schema()
	if s.Kind != "slot" {
		t.Errorf("Schema().Kind = %q, want %q", s.Kind, "slot")
	}
	if len(s.Fields) == 0 {
		t.Error("Schema().Fields is empty")
	}
}

// --- Create ---

func TestSlotCreate(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	rec, err := sr.Create(ctx, "feat-login", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if rec.Kind != "slot" {
		t.Errorf("record Kind = %q, want %q", rec.Kind, "slot")
	}
	if rec.ID != "feat-login" {
		t.Errorf("record ID = %q, want %q", rec.ID, "feat-login")
	}

	// Check copy path exists as a sibling
	copyPath, ok := rec.Fields["path"].(string)
	if !ok || copyPath == "" {
		t.Fatal("record missing path field")
	}
	if _, err := os.Stat(copyPath); err != nil {
		t.Fatalf("copy path does not exist: %v", err)
	}

	// Verify .git exists in copy
	if _, err := os.Stat(filepath.Join(copyPath, ".git")); err != nil {
		t.Fatalf(".git not found in copy: %v", err)
	}

	// Check branch field is the bare slot name
	branch, ok := rec.Fields["branch"].(string)
	if !ok || branch == "" {
		t.Fatal("record missing branch field")
	}
	if branch != "feat-login" {
		t.Errorf("branch = %q, want %q", branch, "feat-login")
	}

	// Verify the branch was actually checked out
	actualBranch := currentBranch(t, copyPath)
	if actualBranch != "feat-login" {
		t.Errorf("actual branch = %q, want %q", actualBranch, "feat-login")
	}

	// Verify cases directory was created
	casesDir := filepath.Join(copyPath, "slots", "feat-login", "cases")
	if _, err := os.Stat(casesDir); err != nil {
		t.Fatalf("cases dir not created: %v", err)
	}
}

// --- Name validation ---

func TestSlotNameValidation(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	cases := []struct {
		name    string
		wantErr bool
	}{
		{"good-name", false},
		{"a", false},
		{"abc123", false},
		{"a-b-c", false},
		{"", true},
		{"123abc", true},
		{"-bad", true},
		{"UPPER", true},
		{"has space", true},
		{"has_underscore", true},
	}
	for _, tc := range cases {
		_, err := sr.Create(ctx, tc.name, nil)
		if tc.wantErr && err == nil {
			t.Errorf("Create(%q): expected error, got nil", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("Create(%q): unexpected error: %v", tc.name, err)
		}
	}
}

// --- ValidateSlotName standalone ---

func TestValidateSlotName(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"good", false},
		{"a-b-c", false},
		{"abc123", false},
		{"", true},
		{"  ", true},
		{"-bad", true},
		{"UPPER", true},
		{"has_underscore", true},
		{"123abc", true},
	}
	for _, tc := range cases {
		err := ValidateSlotName(tc.name)
		if tc.wantErr && err == nil {
			t.Errorf("ValidateSlotName(%q): expected error", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidateSlotName(%q): unexpected error: %v", tc.name, err)
		}
	}
}

// --- List ---

func TestSlotList(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	// Create two slots
	_, err := sr.Create(ctx, "alpha", nil)
	if err != nil {
		t.Fatalf("Create alpha: %v", err)
	}
	_, err = sr.Create(ctx, "beta", nil)
	if err != nil {
		t.Fatalf("Create beta: %v", err)
	}

	records, err := sr.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("List returned %d records, want 2", len(records))
	}

	names := map[string]bool{}
	for _, r := range records {
		names[r.ID] = true
	}
	if !names["alpha"] || !names["beta"] {
		t.Errorf("List names = %v, want alpha and beta", names)
	}
}

// --- Get ---

func TestSlotGet(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	_, err := sr.Create(ctx, "target", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec, err := sr.Get(ctx, "target")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.ID != "target" {
		t.Errorf("Get ID = %q, want %q", rec.ID, "target")
	}

	// Get nonexistent
	_, err = sr.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Get(nonexistent): expected error, got nil")
	}
}

// --- Delete ---

func TestSlotDelete(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	created, err := sr.Create(ctx, "doomed", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	copyPath := created.Fields["path"].(string)

	err = sr.Delete(ctx, "doomed")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Verify copy directory is gone
	if _, err := os.Stat(copyPath); err == nil {
		t.Error("copy path still exists after Delete")
	}

	// Verify it's not in list anymore
	records, err := sr.List(ctx, nil)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, r := range records {
		if r.ID == "doomed" {
			t.Error("deleted slot still appears in List")
		}
	}
}

// --- CopyRepo ---

func TestCopyRepo(t *testing.T) {
	repoDir := setupGitRepo(t)
	fs := &realFS{}
	ex := &realExec{}

	dst := filepath.Join(filepath.Dir(repoDir), "myrepo-copy")

	err := CopyRepo(fs, ex, repoDir, dst)
	if err != nil {
		t.Fatalf("CopyRepo: %v", err)
	}

	// Verify .git exists in copy
	if _, err := os.Stat(filepath.Join(dst, ".git")); err != nil {
		t.Fatalf(".git missing in copy: %v", err)
	}

	// CopyRepo should fail if dst already exists
	err = CopyRepo(fs, ex, repoDir, dst)
	if err == nil {
		t.Error("CopyRepo to existing dst should fail")
	}
}

// --- SlotNameFromPath ---

func TestSlotNameFromPath(t *testing.T) {
	cases := []struct {
		path    string
		prefix  string
		want    string
		wantErr bool
	}{
		{"/tmp/myrepo-alpha", "myrepo", "alpha", false},
		{"/tmp/myrepo-feat-login", "myrepo", "feat-login", false},
		{"/tmp/other-thing", "myrepo", "", true},
		{"/tmp/myrepo-", "myrepo", "", true},
	}
	for _, tc := range cases {
		got, err := SlotNameFromPath(tc.path, tc.prefix)
		if tc.wantErr {
			if err == nil {
				t.Errorf("SlotNameFromPath(%q, %q): expected error", tc.path, tc.prefix)
			}
			continue
		}
		if err != nil {
			t.Errorf("SlotNameFromPath(%q, %q): %v", tc.path, tc.prefix, err)
			continue
		}
		if got != tc.want {
			t.Errorf("SlotNameFromPath(%q, %q) = %q, want %q", tc.path, tc.prefix, got, tc.want)
		}
	}
}

// --- IsDirty ---

func TestIsDirty(t *testing.T) {
	t.Run("clean repo", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		dirty, err := IsDirty(&realExec{}, repoDir)
		if err != nil {
			t.Fatal(err)
		}
		if dirty {
			t.Fatal("fresh repo should not be dirty")
		}
	})

	t.Run("dirty repo", func(t *testing.T) {
		repoDir := setupGitRepo(t)
		os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("x"), 0o644)
		dirty, err := IsDirty(&realExec{}, repoDir)
		if err != nil {
			t.Fatal(err)
		}
		if !dirty {
			t.Fatal("repo with untracked file should be dirty")
		}
	})
}

// --- Config loading (copy-based) ---

func TestConfigLoadCopyBased(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "myrepo")
	os.MkdirAll(filepath.Join(repoRoot, ".agentops"), 0o755)

	cfg, err := LoadSlotConfig(&realFS{}, filepath.Join(repoRoot, ".agentops"), repoRoot)
	if err != nil {
		t.Fatalf("LoadSlotConfig: %v", err)
	}
	if cfg.BaseBranch != "main" {
		t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "main")
	}
	if cfg.CopyPrefix != "myrepo" {
		t.Errorf("CopyPrefix = %q, want %q", cfg.CopyPrefix, "myrepo")
	}
}

func TestConfigLoadFromFile(t *testing.T) {
	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "myrepo")
	agentopsDir := filepath.Join(repoRoot, ".agentops")
	os.MkdirAll(agentopsDir, 0o755)

	yaml := `base_branch: develop
copy_prefix: custom
`
	os.WriteFile(filepath.Join(agentopsDir, "slot.yaml"), []byte(yaml), 0o644)

	cfg, err := LoadSlotConfig(&realFS{}, agentopsDir, repoRoot)
	if err != nil {
		t.Fatalf("LoadSlotConfig: %v", err)
	}
	if cfg.BaseBranch != "develop" {
		t.Errorf("BaseBranch = %q, want %q", cfg.BaseBranch, "develop")
	}
	if cfg.CopyPrefix != "custom" {
		t.Errorf("CopyPrefix = %q, want %q", cfg.CopyPrefix, "custom")
	}
}

func TestConfigCopyPath(t *testing.T) {
	cfg := &SlotConfig{
		CopyPrefix: "myrepo",
		BaseBranch: "main",
	}

	got := cfg.CopyPath("/home/user/repos", "alpha")
	want := "/home/user/repos/myrepo-alpha"
	if got != want {
		t.Errorf("CopyPath = %q, want %q", got, want)
	}
}

// --- CommitsBehind ---

func TestCommitsBehind(t *testing.T) {
	repoDir := setupGitRepo(t)
	ex := &realExec{}

	// Create a branch for testing
	cmd := exec.Command("git", "checkout", "-b", "test-behind")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout: %s %v", out, err)
	}

	// Go back to main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout main: %s %v", out, err)
	}

	// Initially 0 behind
	n, err := CommitsBehind(ex, repoDir, "test-behind", "main")
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 behind, got %d", n)
	}

	// Add commits on main
	for i := 0; i < 3; i++ {
		c := exec.Command("git", "commit", "--allow-empty", "-m", "advance main")
		c.Dir = repoDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("commit: %s %v", out, err)
		}
	}

	n, err = CommitsBehind(ex, repoDir, "test-behind", "main")
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("expected 3 behind, got %d", n)
	}
}

// --- GitError ---

func TestGitError(t *testing.T) {
	ge := &GitError{
		Args:   []string{"status", "--porcelain"},
		Output: "fatal: not a repo",
		Err:    os.ErrNotExist,
	}

	msg := ge.Error()
	if msg != "git status --porcelain: fatal: not a repo" {
		t.Errorf("unexpected error message: %s", msg)
	}
	if ge.Unwrap() != os.ErrNotExist {
		t.Error("Unwrap should return underlying error")
	}
}

// --- Doctor ---

func TestDoctorCleanSlot(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	// Create a clean slot
	_, err := sr.Create(ctx, "clean", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	checks, err := sr.Doctor(ctx)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}

	// Should have exactly one "ok" check for the clean slot
	found := false
	for _, c := range checks {
		if c.Name == "clean" && c.Status == "ok" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ok check for clean slot, got: %+v", checks)
	}
}

func TestDoctorDirtySlot(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	rec, err := sr.Create(ctx, "dirty", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Make the slot dirty
	copyPath := rec.Fields["path"].(string)
	os.WriteFile(filepath.Join(copyPath, "uncommitted.txt"), []byte("dirty"), 0o644)

	checks, err := sr.Doctor(ctx)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}

	found := false
	for _, c := range checks {
		if c.Name == "dirty" && c.Status == "dirty" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dirty check, got: %+v", checks)
	}
}

func TestDoctorBehindBase(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	rec, err := sr.Create(ctx, "behind", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Advance main branch inside the copy so the slot branch falls behind.
	copyPath := rec.Fields["path"].(string)
	advanceCmds := [][]string{
		{"git", "checkout", "main"},
	}
	for i := 0; i < 2; i++ {
		advanceCmds = append(advanceCmds, []string{"git", "commit", "--allow-empty", "-m", "advance main"})
	}
	advanceCmds = append(advanceCmds, []string{"git", "checkout", "behind"})

	for _, args := range advanceCmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = copyPath
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s %v", args, out, err)
		}
	}

	checks, err := sr.Doctor(ctx)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}

	found := false
	for _, c := range checks {
		if c.Name == "behind" && c.Status == "behind" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected behind check, got: %+v", checks)
	}
}

// --- Prune ---

func TestPruneDryRun(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	_, err := sr.Create(ctx, "pruneme", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Dry-run prune (confirm=false)
	results, err := sr.Prune(ctx, false)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Name == "pruneme" && r.Action == "would_remove" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected would_remove for pruneme, got: %+v", results)
	}

	// Verify copy still exists (dry-run)
	records, err := sr.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 slot after dry-run prune, got %d", len(records))
	}
}

func TestPruneConfirm(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	_, err := sr.Create(ctx, "pruneme", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Actual prune (confirm=true)
	results, err := sr.Prune(ctx, true)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Name == "pruneme" && r.Action == "removed" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected removed for pruneme, got: %+v", results)
	}

	// Verify copy is gone
	records, err := sr.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 slots after prune, got %d", len(records))
	}
}

func TestPruneSkipsDirty(t *testing.T) {
	repoDir := setupGitRepo(t)
	sr, ctx := newTestResource(t, repoDir)

	rec, err := sr.Create(ctx, "dirtypruneme", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Make it dirty
	copyPath := rec.Fields["path"].(string)
	os.WriteFile(filepath.Join(copyPath, "uncommitted.txt"), []byte("dirty"), 0o644)

	results, err := sr.Prune(ctx, true)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Name == "dirtypruneme" && r.Action == "skipped" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected skipped for dirty slot, got: %+v", results)
	}

	// Verify copy still exists
	records, err := sr.List(ctx, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 slot after prune (skipped dirty), got %d", len(records))
	}
}

// --- FindRepoRoot ---

func TestFindRepoRoot(t *testing.T) {
	repoDir := setupGitRepo(t)
	fs := &realFS{}

	// From repo root
	root, err := FindRepoRoot(fs, repoDir)
	if err != nil {
		t.Fatalf("FindRepoRoot: %v", err)
	}
	if root != repoDir {
		t.Errorf("FindRepoRoot = %q, want %q", root, repoDir)
	}

	// From subdirectory
	subDir := filepath.Join(repoDir, "subdir")
	os.MkdirAll(subDir, 0o755)
	root, err = FindRepoRoot(fs, subDir)
	if err != nil {
		t.Fatalf("FindRepoRoot from subdir: %v", err)
	}
	if root != repoDir {
		t.Errorf("FindRepoRoot from subdir = %q, want %q", root, repoDir)
	}

	// Non-repo
	_, err = FindRepoRoot(fs, t.TempDir())
	if err == nil {
		t.Error("FindRepoRoot in non-repo should fail")
	}
}

// currentBranch returns the current branch name in the repo.
func currentBranch(t *testing.T, repoDir string) string {
	t.Helper()
	c := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	c.Dir = repoDir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("get current branch: %s %v", out, err)
	}
	branch := string(out)
	for len(branch) > 0 && (branch[len(branch)-1] == '\n' || branch[len(branch)-1] == '\r') {
		branch = branch[:len(branch)-1]
	}
	return branch
}
