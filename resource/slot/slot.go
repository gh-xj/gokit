package slotresource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentops "github.com/gh-xj/agentops"
	"github.com/gh-xj/agentops/dal"
	"github.com/gh-xj/agentops/resource"
)

// SlotResource implements Resource, Deleter, Syncer, Doctor, and Pruner for copy-based slots.
type SlotResource struct {
	fs   dal.FileSystem
	exec dal.Executor
}

// New creates a SlotResource with the given filesystem and executor.
func New(fs dal.FileSystem, exec dal.Executor) *SlotResource {
	return &SlotResource{fs: fs, exec: exec}
}

// Schema returns the resource schema for slots.
func (s *SlotResource) Schema() resource.ResourceSchema {
	return resource.ResourceSchema{
		Kind:        "slot",
		Description: "Copy-based development slot",
		Fields: []resource.FieldDef{
			{Name: "name", Type: "string", Required: true},
			{Name: "path", Type: "string", Required: true},
			{Name: "branch", Type: "string", Required: true},
		},
		CreateArgs: []resource.ArgDef{
			{Name: "name", Description: "Slot name (lowercase alphanumeric with hyphens)", Required: true},
		},
	}
}

// projectDir resolves the project directory from the AppContext or detects it
// from the current git repo root.
func (s *SlotResource) projectDir(ctx *agentops.AppContext) (string, error) {
	if v, ok := ctx.Values["project_dir"]; ok {
		if dir, ok := v.(string); ok && dir != "" {
			return dir, nil
		}
	}
	out, err := s.exec.Run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("detect project dir: %w", err)
	}
	dir := strings.TrimSpace(out)
	if dir == "" {
		return "", fmt.Errorf("not inside a git repository")
	}
	return dir, nil
}

// loadConfig resolves the project directory and loads the slot config.
// The project directory is resolved to its real path (symlinks evaluated) to
// ensure consistent path comparisons.
func (s *SlotResource) loadConfig(ctx *agentops.AppContext) (string, *SlotConfig, error) {
	projectDir, err := s.projectDir(ctx)
	if err != nil {
		return "", nil, err
	}
	// Resolve symlinks (e.g., /var -> /private/var on macOS) so paths
	// are consistent.
	if resolved, err := os.Readlink(projectDir); err == nil {
		projectDir = resolved
	}
	if resolved, err := filepath.EvalSymlinks(projectDir); err == nil {
		projectDir = resolved
	}
	agentopsDir := filepath.Join(projectDir, ".agentops")
	cfg, err := LoadSlotConfig(s.fs, agentopsDir, projectDir)
	if err != nil {
		return "", nil, fmt.Errorf("load slot config: %w", err)
	}
	return projectDir, cfg, nil
}

// slotInfo holds parsed information about a slot.
type slotInfo struct {
	Name   string
	Path   string
	Branch string
}

// Create validates the name, copies the repo, and checks out a new branch.
func (s *SlotResource) Create(ctx *agentops.AppContext, slug string, opts map[string]string) (*resource.Record, error) {
	if err := ValidateSlotName(slug); err != nil {
		return nil, err
	}

	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	parentDir := filepath.Dir(projectDir)
	copyPath := cfg.CopyPath(parentDir, slug)

	// Copy the repo
	if err := CopyRepo(s.fs, s.exec, projectDir, copyPath); err != nil {
		return nil, fmt.Errorf("create slot %q: %w", slug, err)
	}

	// Checkout a new branch named after the slot
	if err := CheckoutNewBranch(s.exec, copyPath, slug); err != nil {
		// Clean up on failure
		os.RemoveAll(copyPath)
		return nil, fmt.Errorf("create slot %q: checkout branch: %w", slug, err)
	}

	// Create the cases directory
	casesDir := filepath.Join(copyPath, "slots", slug, "cases")
	if err := s.fs.EnsureDir(casesDir); err != nil {
		// Clean up on failure
		os.RemoveAll(copyPath)
		return nil, fmt.Errorf("create slot %q: create cases dir: %w", slug, err)
	}

	info := slotInfo{
		Name:   slug,
		Path:   copyPath,
		Branch: slug,
	}
	return infoToRecord(info), nil
}

// List returns all slots for the project by scanning sibling directories.
func (s *SlotResource) List(ctx *agentops.AppContext, filter resource.Filter) ([]resource.Record, error) {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	infos, err := s.listCopies(projectDir, cfg)
	if err != nil {
		return nil, err
	}

	records := make([]resource.Record, 0, len(infos))
	for _, info := range infos {
		records = append(records, *infoToRecord(info))
	}
	return records, nil
}

// Get returns a single slot by name.
func (s *SlotResource) Get(ctx *agentops.AppContext, id string) (*resource.Record, error) {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	infos, err := s.listCopies(projectDir, cfg)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.Name == id {
			return infoToRecord(info), nil
		}
	}
	return nil, fmt.Errorf("slot %q not found", id)
}

// Delete removes a slot copy after checking for uncommitted changes.
func (s *SlotResource) Delete(ctx *agentops.AppContext, id string) error {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(projectDir)
	copyPath := cfg.CopyPath(parentDir, id)

	if !s.fs.Exists(copyPath) {
		return fmt.Errorf("slot %q not found at %s", id, copyPath)
	}

	// Check for uncommitted changes
	dirty, err := IsDirty(s.exec, copyPath)
	if err != nil {
		return fmt.Errorf("check dirty: %w", err)
	}
	if dirty {
		return fmt.Errorf("slot %q has uncommitted changes; commit or stash first", id)
	}

	return os.RemoveAll(copyPath)
}

// Sync rebases the slot branch onto origin/<base_branch>.
func (s *SlotResource) Sync(ctx *agentops.AppContext, id string) error {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return err
	}

	parentDir := filepath.Dir(projectDir)
	copyPath := cfg.CopyPath(parentDir, id)

	if !s.fs.Exists(copyPath) {
		return fmt.Errorf("slot %q not found at %s", id, copyPath)
	}

	return FetchAndRebase(s.exec, copyPath, cfg.BaseBranch)
}

// Doctor runs health checks on all active slot copies.
func (s *SlotResource) Doctor(ctx *agentops.AppContext) ([]resource.DoctorCheck, error) {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	infos, err := s.listCopies(projectDir, cfg)
	if err != nil {
		return nil, err
	}

	var results []resource.DoctorCheck

	for _, info := range infos {
		slotHasIssue := false

		// Check 1: Missing .git
		gitPath := filepath.Join(info.Path, ".git")
		if !s.fs.Exists(gitPath) {
			slotHasIssue = true
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "missing_git",
				Message:  ".git not found",
				Severity: "err",
			})
		}

		// Check 2: Dirty working tree
		dirty, dirtyErr := IsDirty(s.exec, info.Path)
		if dirtyErr != nil {
			slotHasIssue = true
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "check_error",
				Message:  fmt.Sprintf("cannot check dirty status: %v", dirtyErr),
				Severity: "warn",
			})
		} else if dirty {
			slotHasIssue = true
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "dirty",
				Message:  "uncommitted changes",
				Severity: "warn",
			})
		}

		// Check 3: Behind base branch
		behind, behindErr := CommitsBehind(s.exec, info.Path, info.Branch, cfg.BaseBranch)
		if behindErr != nil {
			slotHasIssue = true
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "check_error",
				Message:  fmt.Sprintf("cannot check behind status: %v", behindErr),
				Severity: "warn",
			})
		} else if behind > 0 {
			slotHasIssue = true
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "behind",
				Message:  fmt.Sprintf("%d commits behind %s", behind, cfg.BaseBranch),
				Severity: "warn",
			})
		}

		if !slotHasIssue {
			results = append(results, resource.DoctorCheck{
				Name:     info.Name,
				Status:   "ok",
				Message:  "clean, up to date",
				Severity: "ok",
			})
		}
	}

	return results, nil
}

// Prune removes clean stale copies. Dry-run by default (confirm=false).
func (s *SlotResource) Prune(ctx *agentops.AppContext, confirm bool) ([]resource.PruneResult, error) {
	projectDir, cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	infos, err := s.listCopies(projectDir, cfg)
	if err != nil {
		return nil, err
	}

	var results []resource.PruneResult

	for _, info := range infos {
		dirty, dirtyErr := IsDirty(s.exec, info.Path)
		if dirtyErr != nil {
			results = append(results, resource.PruneResult{
				Name:   info.Name,
				Path:   info.Path,
				Action: "skipped",
				Reason: fmt.Sprintf("cannot check dirty status: %v", dirtyErr),
			})
			continue
		}

		if dirty {
			results = append(results, resource.PruneResult{
				Name:   info.Name,
				Path:   info.Path,
				Action: "skipped",
				Reason: "dirty (uncommitted changes)",
			})
			continue
		}

		// Clean slot — remove or report
		if confirm {
			if err := os.RemoveAll(info.Path); err != nil {
				results = append(results, resource.PruneResult{
					Name:   info.Name,
					Path:   info.Path,
					Action: "skipped",
					Reason: fmt.Sprintf("remove failed: %v", err),
				})
				continue
			}
			results = append(results, resource.PruneResult{
				Name:   info.Name,
				Path:   info.Path,
				Action: "removed",
			})
		} else {
			results = append(results, resource.PruneResult{
				Name:   info.Name,
				Path:   info.Path,
				Action: "would_remove",
			})
		}
	}

	return results, nil
}

// listCopies scans the parent directory for slot copies matching the prefix.
func (s *SlotResource) listCopies(projectDir string, cfg *SlotConfig) ([]slotInfo, error) {
	parentDir := filepath.Dir(projectDir)
	entries, err := s.fs.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("read parent dir: %w", err)
	}

	prefix := cfg.CopyPrefix + "-"
	var slots []slotInfo

	for _, entry := range entries {
		if !entry.IsDir {
			continue
		}
		if !strings.HasPrefix(entry.Name, prefix) {
			continue
		}
		name := strings.TrimPrefix(entry.Name, prefix)
		if name == "" {
			continue
		}

		copyPath := filepath.Join(parentDir, entry.Name)

		// Verify .git exists to confirm it's a valid repo copy
		gitPath := filepath.Join(copyPath, ".git")
		if !s.fs.Exists(gitPath) {
			continue
		}

		// Get current branch
		branch, err := CurrentBranch(s.exec, copyPath)
		if err != nil {
			branch = name // fallback to name
		}

		slots = append(slots, slotInfo{
			Name:   name,
			Path:   copyPath,
			Branch: branch,
		})
	}

	return slots, nil
}

// infoToRecord converts a slotInfo to a resource.Record.
func infoToRecord(info slotInfo) *resource.Record {
	return &resource.Record{
		Kind: "slot",
		ID:   info.Name,
		Fields: map[string]any{
			"name":   info.Name,
			"path":   info.Path,
			"branch": info.Branch,
		},
		RawPath: info.Path,
	}
}
