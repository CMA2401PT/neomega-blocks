//go:build resourcegen

package blockconverts

import (
	"encoding/json"
	"neomega-blocks/ir"
	v1 "neomega-blocks/presets/block-converts/v1"
	"neomega-blocks/storage/num_lookup"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestSchematic(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	schematicFile := filepath.Join(filepath.Dir(cur), "schematic.json")
	bs, err := os.ReadFile(schematicFile)
	if err != nil {
		panic(err)
	}
	var schematicData map[string]string
	err = json.Unmarshal(bs, &schematicData)
	if err != nil {
		panic(err)
	}

	lookup := num_lookup.NewTwoLayerNumMap[uint32, uint32]()
	for schematicIDStr, javaStr := range schematicData {
		blk := ir.IRBlock(javaStr)
		if !blk.IsValid() {
			t.Fatalf("%v %v", schematicIDStr, javaStr)
		}
		blk = blk.Tidy()
		bedrockBlock, ok := v1.ConvertFromJava(blk)
		if !ok {
			t.Fatalf("%v %v", schematicIDStr, javaStr)
		}
		// fmt.Println(javaStr, " -> ", bedrockBlock)
		schematicNamePair := strings.Split(schematicIDStr, ":")
		if len(schematicNamePair) != 2 {
			panic("invalid schematic id")
		}
		schematicID, err := strconv.Atoi(schematicNamePair[0])
		if err != nil {
			panic(err)
		}
		schematicMeta, err := strconv.Atoi(schematicNamePair[1])
		if err != nil {
			panic(err)
		}
		schematicHash := uint32(schematicID)<<16 | uint32(schematicMeta)
		lookup.Set(schematicHash, bedrockBlock.Hash())
	}
	if err := os.WriteFile("../../../presets/block-converts/v1/6_schematic_fan_in.bin", lookup.Dump(), 0o644); err != nil {
		t.Fatalf("write defaults bin: %v", err)
	}
}
