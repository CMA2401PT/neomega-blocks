package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	presets "neomega-blocks/presets/block-converts/v1"
)

// This read-only check joins the real palette, the runtime converter and a
// generated resource bundle. It deliberately has no dependency on the builder.
func TestJavaTargetAssets(t *testing.T) {
	dir := os.Getenv("NEOMEGA_ASSETS_DIR")
	if dir == "" {
		t.Skip("set NEOMEGA_ASSETS_DIR to the final resource directory")
	}
	data, err := os.ReadFile(filepath.Join(dir, "build-info.json"))
	if err != nil {
		t.Fatal(err)
	}
	var info struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	if info.Version != presets.JavaTargetVersion {
		t.Fatalf("resources %s differ from conversion target %s", info.Version, presets.JavaTargetVersion)
	}
	data, err = os.ReadFile(filepath.Join(dir, "blockstates.json"))
	if err != nil {
		t.Fatal(err)
	}
	var states map[string]struct {
		Variants  map[string]json.RawMessage `json:"variants"`
		Multipart []json.RawMessage          `json:"multipart"`
	}
	if err := json.Unmarshal(data, &states); err != nil {
		t.Fatal(err)
	}
	palette := ReadBlocks("block_palette.json")
	names := map[string]bool{}
	for _, block := range palette {
		source, ok := presets.BedrockHashToBlockStrict(block.IRBlock.Hash())
		if !ok {
			t.Fatalf("palette hash missing: %s", block.IRBlock)
		}
		java, ok := presets.ConvertToJava(source)
		if !ok {
			t.Fatalf("Java conversion missing: %s", source)
		}
		name := strings.TrimPrefix(java.NameWithNameSpace(), "minecraft:")
		state, ok := states[name]
		if !ok {
			t.Fatalf("converted block has no resource: %s -> %s", source, java)
		}
		names[name] = true
		if len(state.Variants) == 0 {
			if len(state.Multipart) == 0 {
				t.Fatalf("resource has no model choices: %s", name)
			}
			continue
		}
		// PyMCTranslate represents water as an extra block, outside this
		// single-layer palette. Check the dry static preview selection.
		properties := map[string]string{"waterlogged": "false"}
		for key, value := range java.States().KeyValues {
			properties[key] = fmt.Sprint(value)
		}
		found := false
		for selector := range state.Variants {
			matches := true
			if selector != "" {
				for _, pair := range strings.Split(selector, ",") {
					key, values, _ := strings.Cut(pair, "=")
					allowed := false
					for _, value := range strings.Split(values, "|") {
						allowed = allowed || properties[key] == value
					}
					matches = matches && allowed
				}
			}
			found = found || matches
		}
		if !found {
			t.Fatalf("converted state matches no variant: %s -> %s", source, java)
		}
	}
	t.Logf("checked %d palette states, %d Java block names against resources %s", len(palette), len(names), info.Version)
}
