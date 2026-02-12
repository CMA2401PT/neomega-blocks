package scripts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"neomega-blocks/ir"
)

func TestReadDefaultsAllJSON(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	defaultsPath := filepath.Join(
		filepath.Dir(cur),
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

	files, err := filepath.Glob(filepath.Join(defaultsPath, "*.json"))
	if err != nil {
		t.Fatalf("glob defaults jsons: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no defaults json files found")
	}

	defaults := ReadDefaults(defaultsPath)
	if len(defaults) != len(files) {
		t.Fatalf("processed file count mismatch: got %d, want %d", len(defaults), len(files))
	}

	for _, f := range files {
		base := strings.TrimSuffix(filepath.Base(f), ".json")
		name := "minecraft:" + base
		states, found := defaults[name]
		if !found {
			t.Fatalf("missing defaults for %s", name)
		}
		if !states.IsTidy() {
			t.Fatalf("states for %s is not tidy: %q", name, states)
		}

		payload, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var doc struct {
			Defaults map[string]string `json:"defaults"`
		}
		if err := json.Unmarshal(payload, &doc); err != nil {
			t.Fatalf("decode %s: %v", f, err)
		}
		gotCount := 0
		for range states.KeyValueStrings {
			gotCount++
		}
		if gotCount != len(doc.Defaults) {
			t.Fatalf("defaults count mismatch for %s: got %d, want %d", name, gotCount, len(doc.Defaults))
		}
	}

	if got := defaults["minecraft:air"]; got != ir.IRStates("[]") {
		t.Fatalf(`defaults["minecraft:air"] = %q, want []`, got)
	}
	if got := defaults["minecraft:acacia_door"]; got != ir.IRStates(`["direction"=0,"door_hinge_bit"=false,"open_bit"=false,"upper_block_bit"=false]`) {
		t.Fatalf(`defaults["minecraft:acacia_door"] = %q`, got)
	}
	if got := defaults["minecraft:campfire"]; got != ir.IRStates(`["extinguished"=false,"minecraft:cardinal_direction"="south"]`) {
		t.Fatalf(`defaults["minecraft:campfire"] = %q`, got)
	}
}
