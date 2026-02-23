package harnessloop

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func CurrentBranch(repoRoot string) string {
	out, err := runGit(repoRoot, "branch", "--show-current")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func EnsureBranch(repoRoot, branch string) error {
	_, err := runGit(repoRoot, "checkout", "-B", branch)
	return err
}

func CommitIfDirty(repoRoot, msg string) (bool, error) {
	status, err := runGit(repoRoot, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(status) == "" {
		return false, nil
	}
	if _, err := runGit(repoRoot, "add", "-A"); err != nil {
		return false, err
	}
	if _, err := runGit(repoRoot, "commit", "-m", msg); err != nil {
		return false, err
	}
	return true, nil
}

func runGit(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out.String())
	}
	return out.String(), nil
}
