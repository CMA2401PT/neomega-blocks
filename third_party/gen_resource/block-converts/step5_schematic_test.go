//go:build resourcegen

package blockconverts

import (
	"compress/gzip"
	"encoding/json"
	"neomega-blocks/ir"
	v1 "neomega-blocks/presets/block-converts/v1"
	java "neomega-blocks/presets/java/v1"
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

	javaMapping := make(map[string]string, len(schematicData))
	lookup := num_lookup.NewTwoLayerNumMap[uint32, uint32]()
	for schematicIDStr, javaStr := range schematicData {
		blk := ir.IRBlock(javaStr)
		if !blk.IsValid() {
			t.Fatalf("%v %v", schematicIDStr, javaStr)
		}
		blk = blk.Tidy()
		// 源表包含 grass、sign 等历史名称，对齐转换中心的 Java 侧，避免经过 Bedrock。
		canonical, ok := v1.JavaCanon(blk)
		if !ok {
			t.Fatalf("Java defaults missing: %s", blk)
		}
		singleLayer := ir.IRBlock(canonical.NameWithNameSpace() + " " + string(canonical.States().Delete("waterlogged"))).Tidy()
		target, ok := java.NormalizeSingleLayer(singleLayer)
		if !ok {
			t.Fatalf("Java target missing: %s", canonical)
		}
		javaMapping[schematicIDStr] = string(target)
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
	// Java 原理图使用对齐目标版本后的直接映射，避免经过 Bedrock 状态集丢掉 Java 属性。
	file, err := os.Create("../../../presets/block-converts/v1/7_schematic_to_java.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	zipped := gzip.NewWriter(file)
	if err = json.NewEncoder(zipped).Encode(javaMapping); err != nil {
		t.Fatal(err)
	}
	if err = zipped.Close(); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}

}
