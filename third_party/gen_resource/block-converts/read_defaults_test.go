//go:build resourcegen

package blockconverts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"neomega-blocks/ir"
)

func TestReadDefaultsAllJSON(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	defaultsPath := filepath.Join(filepath.Dir(cur), "tmp", "defaults", "bedrock_1_21_50.json")
	payload, err := os.ReadFile(defaultsPath)
	if err != nil {
		t.Fatal(err)
	}
	var exported map[string]map[string]string
	if err := json.Unmarshal(payload, &exported); err != nil {
		t.Fatal(err)
	}
	defaults := ReadDefaults(defaultsPath)
	if len(defaults) == 0 || len(defaults) != len(exported) {
		t.Fatalf("processed block count mismatch: got %d, exported %d", len(defaults), len(exported))
	}
	for name, states := range defaults {
		if !states.IsTidy() {
			t.Fatalf("states for %s is not tidy: %q", name, states)
		}

		gotCount := 0
		for range states.KeyValueStrings {
			gotCount++
		}
		if gotCount != len(exported[name]) {
			t.Fatalf("defaults count mismatch for %s: got %d, want %d", name, gotCount, len(exported[name]))
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
