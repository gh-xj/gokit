package slotresource

import (
	"fmt"
	"path/filepath"

	"github.com/gh-xj/agentops/dal"
	"gopkg.in/yaml.v3"
)

// SlotConfig represents the parsed .agentops/slot.yaml configuration.
type SlotConfig struct {
	BaseBranch string `yaml:"base_branch"`
	CopyPrefix string `yaml:"copy_prefix"`
}

// LoadSlotConfig reads and parses .agentops/slot.yaml. If the file does not
// exist, sensible defaults are returned. The repoRoot is used to derive the
// default copy prefix from the repository directory name.
func LoadSlotConfig(fs dal.FileSystem, agentopsDir string, repoRoot string) (*SlotConfig, error) {
	cfg := &SlotConfig{}

	path := filepath.Join(agentopsDir, "slot.yaml")
	data, err := fs.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse slot.yaml: %w", err)
		}
	}
	// If file doesn't exist, use all defaults

	// Apply defaults
	if cfg.BaseBranch == "" {
		cfg.BaseBranch = "main"
	}
	if cfg.CopyPrefix == "" {
		cfg.CopyPrefix = filepath.Base(repoRoot)
	}

	return cfg, nil
}

// CopyPath computes the path for a copy-based slot directory. Copies are placed
// as siblings to the source repo: <parentDir>/<prefix>-<name>.
func (c *SlotConfig) CopyPath(parentDir, name string) string {
	return filepath.Join(parentDir, c.CopyPrefix+"-"+name)
}
