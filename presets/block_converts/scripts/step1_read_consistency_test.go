package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

// 从 PyMCTranslate 的数据库中读取所有方块的默认状态, 并比对和 NEMC 方块的匹配程度
func TestReadDefaultsReadBlocksConsistency(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	baseDir := filepath.Dir(cur)
	palettePath := filepath.Join(baseDir, "block_palette.json")
	defaultsPath := filepath.Join(
		baseDir,
		"PyMCTranslate",
		"json",
		"versions",
		"bedrock_1_21_50",
		"block",
		"blockstate",
		"specification",
		"minecraft",
		"vanilla",
	)

	blocks := ReadBlocks(palettePath)
	defaults := ReadDefaults(defaultsPath)

	type classInfo struct {
		firstKeys  map[string]struct{}
		allKeys    map[string]struct{}
		consistent bool
	}
	classes := map[string]*classInfo{}
	for _, b := range blocks {
		name := b.IRBlock.NameWithNameSpace()
		keys := keySetFromStates(b.IRBlock.States())
		info, found := classes[name]
		if !found {
			classes[name] = &classInfo{
				firstKeys:  cloneSet(keys),
				allKeys:    cloneSet(keys),
				consistent: true,
			}
			continue
		}
		if !equalSet(info.firstKeys, keys) {
			info.consistent = false
		}
		for k := range keys {
			info.allKeys[k] = struct{}{}
		}
	}

	onlyInBlocks := make([]string, 0)
	for name := range classes {
		if _, ok := defaults[name]; !ok {
			onlyInBlocks = append(onlyInBlocks, name)
		}
	}
	sort.Strings(onlyInBlocks)

	onlyInDefaults := make([]string, 0)
	for name := range defaults {
		if _, ok := classes[name]; !ok {
			onlyInDefaults = append(onlyInDefaults, name)
		}
	}
	sort.Strings(onlyInDefaults)

	internalInconsistent := make([]string, 0)
	for name, info := range classes {
		if !info.consistent {
			internalInconsistent = append(internalInconsistent, name)
		}
	}
	sort.Strings(internalInconsistent)

	keyMismatches := make([]string, 0)
	for name, info := range classes {
		ds, ok := defaults[name]
		if !ok {
			continue
		}
		defaultKeys := keySetFromStates(ds)
		if !equalSet(info.allKeys, defaultKeys) {
			keyMismatches = append(keyMismatches, fmt.Sprintf(
				"%s | blocks=%v | defaults=%v",
				name,
				sortedKeys(info.allKeys),
				sortedKeys(defaultKeys),
			))
		}
	}
	sort.Strings(keyMismatches)

	t.Logf("summary: block classes=%d, default classes=%d", len(classes), len(defaults))
	t.Logf("only in read_blocks: %d", len(onlyInBlocks))
	for _, name := range onlyInBlocks {
		t.Logf("\tonly_in_read_blocks: %s", name)
	}
	t.Logf("only in read_defaults: %d", len(onlyInDefaults))
	for _, name := range onlyInDefaults {
		t.Logf("\tonly_in_read_defaults: %s", name)
	}
	t.Logf("read_blocks internal key-set inconsistent classes: %d", len(internalInconsistent))
	for _, name := range internalInconsistent {
		t.Logf("\tinternal_inconsistent_in_read_blocks: %s", name)
	}
	t.Logf("key mismatches between read_blocks and read_defaults: %d", len(keyMismatches))
	for _, line := range keyMismatches {
		t.Logf("\tkey_mismatch: %s", line)
	}
	if len(keyMismatches) > 0 {
		t.Fatalf("stop because key mismatches between read_blocks and read_defaults: %d", len(keyMismatches))
	}

	// 保存 defaults map[string]ir.IRStates，和保存 map[string]string 一样直接 JSON 序列化。
	data, err := json.Marshal(defaults)
	if err != nil {
		t.Fatalf("marshal defaults: %v", err)
	}
	if err := os.WriteFile("../v1/1_bedrock_default_props.bin", data, 0o644); err != nil {
		t.Fatalf("write defaults bin: %v", err)
	}
}

func keySetFromStates(states interface {
	KeyValueStrings(func(string, string) bool)
}) map[string]struct{} {
	out := map[string]struct{}{}
	states.KeyValueStrings(func(key, _ string) bool {
		out[key] = struct{}{}
		return true
	})
	return out
}

func cloneSet(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{}, len(src))
	for k := range src {
		dst[k] = struct{}{}
	}
	return dst
}

func equalSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
