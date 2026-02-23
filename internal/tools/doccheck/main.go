package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	repoRoot := flag.String("repo-root", ".", "repository root")
	flag.Parse()

	if err := run(*repoRoot); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("doc drift check passed")
}

func run(repoRoot string) error {
	helpOut, err := cliHelp(repoRoot)
	if err != nil {
		return err
	}
	leanSig, labSig, err := extractLoopSignatures(helpOut)
	if err != nil {
		return err
	}

	targets := []string{
		filepath.Join(repoRoot, "skills", "verification-loop", "SKILL.md"),
		filepath.Join(repoRoot, "skills", "verification-loop", "README.md"),
	}

	missing := []string{}
	for _, path := range targets {
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		text := string(raw)
		if !strings.Contains(text, leanSig) {
			missing = append(missing, fmt.Sprintf("%s missing '%s'", path, leanSig))
		}
		if !strings.Contains(text, labSig) {
			missing = append(missing, fmt.Sprintf("%s missing '%s'", path, labSig))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("skill docs drift detected:\n- %s", strings.Join(missing, "\n- "))
	}
	return nil
}

func cliHelp(repoRoot string) (string, error) {
	cmd := exec.Command("go", "run", "./cmd/agentcli", "--help")
	cmd.Dir = repoRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run help command: %w\n%s", err, out.String())
	}
	return out.String(), nil
}

func extractLoopSignatures(helpText string) (string, string, error) {
	leanRe := regexp.MustCompile(`agentcli loop \[[^\]]+\]`)
	labRe := regexp.MustCompile(`agentcli loop lab \[[^\]]+\]`)
	lean := strings.TrimSpace(leanRe.FindString(helpText))
	lab := strings.TrimSpace(labRe.FindString(helpText))
	if lean == "" || lab == "" {
		return "", "", fmt.Errorf("could not extract loop command signatures from CLI help")
	}
	return lean, lab, nil
}
