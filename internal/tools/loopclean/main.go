package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	repoRoot := flag.String("repo-root", ".", "repo root")
	keepCompare := flag.Int("keep-compare", 20, "keep latest compare markdown files")
	flag.Parse()

	if err := clean(*repoRoot, *keepCompare); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("loop clean completed")
}

func clean(repoRoot string, keepCompare int) error {
	if keepCompare < 0 {
		keepCompare = 0
	}
	compareDir := filepath.Join(repoRoot, ".docs", "onboarding-loop", "compare")
	entries, err := os.ReadDir(compareDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) > keepCompare {
		for _, name := range files[:len(files)-keepCompare] {
			if err := os.Remove(filepath.Join(compareDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}
