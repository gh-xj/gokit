package slotresource

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gh-xj/agentops/dal"
)

var slotNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// GitError wraps a failed git command with its arguments and output.
type GitError struct {
	Args   []string
	Output string
	Err    error
}

func (e *GitError) Error() string {
	return "git " + strings.Join(e.Args, " ") + ": " + e.Output
}

func (e *GitError) Unwrap() error { return e.Err }

// gitRun executes a git command in the given directory via dal.Executor.
func gitRun(exec dal.Executor, dir string, args ...string) (string, error) {
	out, err := exec.RunInDir(dir, "git", args...)
	if err != nil {
		return "", &GitError{Args: args, Output: strings.TrimSpace(out), Err: err}
	}
	return out, nil
}

// ValidateSlotName checks whether name is a valid slot name.
// Names must match ^[a-z][a-z0-9-]*$ and must not be empty or whitespace.
func ValidateSlotName(name string) error {
	if name == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("slot name must not be empty")
	}
	if !slotNamePattern.MatchString(name) {
		return fmt.Errorf("invalid slot name %q: must match ^[a-z][a-z0-9-]*$", name)
	}
	return nil
}

// CopyRepo copies the source directory to dst using cp -r.
// Returns an error if dst already exists.
func CopyRepo(fs dal.FileSystem, exec dal.Executor, src, dst string) error {
	if fs.Exists(dst) {
		return fmt.Errorf("destination already exists: %s", dst)
	}
	_, err := exec.Run("cp", "-r", src, dst)
	if err != nil {
		return fmt.Errorf("cp -r %s %s: %w", src, dst, err)
	}
	return nil
}

// SlotNameFromPath extracts a slot name from a directory name given the prefix.
// For a dir named "<prefix>-<name>", returns "<name>".
func SlotNameFromPath(path, prefix string) (string, error) {
	base := filepath.Base(path)
	full := prefix + "-"
	if !strings.HasPrefix(base, full) {
		return "", fmt.Errorf("path %q does not match prefix %q", base, prefix)
	}
	name := strings.TrimPrefix(base, full)
	if name == "" {
		return "", fmt.Errorf("path %q has empty slot name after prefix", base)
	}
	return name, nil
}

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(exec dal.Executor, dir string) (bool, error) {
	out, err := gitRun(exec, dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// CheckoutNewBranch creates and checks out a new branch in the given directory.
func CheckoutNewBranch(exec dal.Executor, dir, branch string) error {
	_, err := gitRun(exec, dir, "checkout", "-b", branch)
	return err
}

// FetchAndRebase fetches the latest base branch and rebases the current branch onto it.
// If rebase conflicts, it aborts and returns an error.
func FetchAndRebase(exec dal.Executor, dir, baseBranch string) error {
	// Fetch latest base branch (ignore errors for local-only repos)
	exec.RunInDir(dir, "git", "fetch", "origin", baseBranch)

	// Rebase
	out, err := exec.RunInDir(dir, "git", "rebase", "origin/"+baseBranch)
	if err != nil {
		// Abort failed rebase
		exec.RunInDir(dir, "git", "rebase", "--abort")
		return fmt.Errorf("rebase failed (aborted): %s: %w", strings.TrimSpace(out), err)
	}
	return nil
}

// CurrentBranch returns the current branch name of the repository at dir.
func CurrentBranch(exec dal.Executor, dir string) (string, error) {
	out, err := gitRun(exec, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// FindRepoRoot walks up from dir looking for a .git directory or file.
func FindRepoRoot(fs dal.FileSystem, dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	dir = absDir

	for {
		gitPath := filepath.Join(dir, ".git")
		if fs.Exists(gitPath) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no git repository found (searched up from %s)", absDir)
		}
		dir = parent
	}
}

// CommitsBehind returns how many commits branch is behind baseBranch.
func CommitsBehind(exec dal.Executor, repoDir, branch, baseBranch string) (int, error) {
	out, err := gitRun(exec, repoDir, "rev-list", "--count", branch+".."+baseBranch)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, fmt.Errorf("parse rev-list count: %w", err)
	}
	return n, nil
}
