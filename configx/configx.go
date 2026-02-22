package configx

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Options controls deterministic config loading precedence.
// Precedence order: Defaults < File < Env < Flags.
type Options struct {
	Defaults map[string]any
	FilePath string
	Env      map[string]string
	Flags    map[string]string
}

// Load merges config sources with deterministic precedence.
func Load(opts Options) (map[string]any, error) {
	out := cloneMap(opts.Defaults)

	if strings.TrimSpace(opts.FilePath) != "" {
		fileVals, err := readObjectFile(opts.FilePath)
		if err != nil {
			return nil, err
		}
		mergeAny(out, fileVals)
	}
	mergeStringMap(out, opts.Env)
	mergeStringMap(out, opts.Flags)
	return out, nil
}

// Decode converts a merged config object into a typed config struct.
func Decode[T any](raw map[string]any) (T, error) {
	var out T
	b, err := json.Marshal(raw)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return out, err
	}
	return out, nil
}

// NormalizeEnv extracts prefixed env vars into config keys.
// Example: prefix APP_, key APP_LOG_LEVEL -> log_level.
func NormalizeEnv(prefix string, environ []string) map[string]string {
	out := map[string]string{}
	for _, pair := range environ {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		key := strings.ToLower(strings.TrimPrefix(k, prefix))
		key = strings.ReplaceAll(key, "__", ".")
		out[key] = v
	}
	return out
}

func readObjectFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func cloneMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeAny(dst, src map[string]any) {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		dst[k] = src[k]
	}
}

func mergeStringMap(dst map[string]any, src map[string]string) {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		dst[k] = src[k]
	}
}
