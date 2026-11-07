package v1_test

import (
	"neomega-blocks/ir"
	java "neomega-blocks/presets/java/v1"
	"slices"
	"testing"
)

func TestBlockToHash(t *testing.T) {
	input := ir.IRBlock(`minecraft:oak_fence ["west"="false","south"="false","north"="true","east"="false"]`)
	want := ir.IRBlock(`minecraft:oak_fence ["east"="false","north"="true","south"="false","west"="false"]`)
	hash := java.BlockToHash(input)
	if hash != want.Hash() {
		t.Fatalf("hash = %d, want %d", hash, want.Hash())
	}
	if got, ok := java.HashToBlock(hash); !ok || got != want {
		t.Fatalf("round trip = (%q, %v), want %q", got, ok, want)
	}
}

func TestHashToBlock(t *testing.T) {
	want := ir.IRBlock(`minecraft:oak_fence ["east"="true","north"="false","south"="true","west"="false"]`)
	if got, ok := java.HashToBlock(want.Hash()); !ok || got != want {
		t.Fatalf("lookup = (%q, %v), want %q", got, ok, want)
	}
	if _, ok := java.HashToBlock(ir.IRBlock("minecraft:unknown []").Hash()); ok {
		t.Fatal("unknown hash should not be found")
	}
}

func TestBlockProperties(t *testing.T) {
	properties, ok := java.BlockProperties("minecraft:oak_slab")
	if !ok {
		t.Fatal("oak_slab should be found")
	}
	property := properties["type"]
	if property.Default != `"bottom"` || !slices.Equal(property.Values, []string{`"top"`, `"bottom"`, `"double"`}) {
		t.Fatalf("type = %+v, want bottom default and top, bottom, double domain", property)
	}
	block := ir.IRBlock(`minecraft:oak_slab ["type"=` + property.Default + `]`)
	if got, ok := java.HashToBlock(java.BlockToHash(block)); !ok || got != block {
		t.Fatalf("default state lookup = (%q, %v), want %q", got, ok, block)
	}
	if properties, ok := java.BlockProperties("air"); !ok || len(properties) != 0 {
		t.Fatalf("air properties = (%v, %v), want empty properties", properties, ok)
	}
	if _, ok := java.BlockProperties("minecraft:unknown"); ok {
		t.Fatal("unknown block should not be found")
	}
}

func TestHashToBlockWaterlogged(t *testing.T) {
	for _, value := range []string{"true", "false"} {
		want := ir.IRBlock(`minecraft:oak_stairs ["facing"="north","half"="bottom","shape"="straight","waterlogged"="` + value + `"]`).Tidy()
		if got, ok := java.HashToBlock(java.BlockToHash(want)); !ok || got != want {
			t.Fatalf("lookup = (%q, %v), want %q", got, ok, want)
		}
	}
}

func TestJavaFullStates(t *testing.T) {
	count := 0
	for block := range java.JavaFullStates() {
		if !block.IsTidy() {
			t.Fatalf("state is not tidy: %q", block)
		}
		count++
	}
	if count != 22398 {
		t.Fatalf("state count = %d, want 22398", count)
	}
}

func TestLegacyStateMigration(t *testing.T) {
	// 完整旧状态先按转换中心解释，再迁移名称和属性，含水仍由独立层表达。
	for source, want := range map[ir.IRBlock]ir.IRBlock{
		`minecraft:chain ["axis"="y"]`:                                             `minecraft:iron_chain ["axis"="y"]`,
		`minecraft:creaking_heart ["active"="true","axis"="x","natural"="true"]`:   `minecraft:creaking_heart ["axis"="x","creaking_heart_state"="awake","natural"="true"]`,
		`minecraft:creaking_heart ["active"="false","axis"="z","natural"="false"]`: `minecraft:creaking_heart ["axis"="z","creaking_heart_state"="uprooted","natural"="false"]`,
	} {
		if got, ok := java.NormalizeSingleLayer(source); !ok || got != want {
			t.Fatalf("%s → %s (%v)，期望 %s", source, got, ok, want)
		}
	}
	count := 0
	for range java.JavaCompatibility() {
		count++
	}
	if count != 2077 {
		t.Fatalf("兼容映射数 = %d，期望 2077", count)
	}
}

func TestTargetBlockProperties(t *testing.T) {
	// 浏览与 fixer 使用当前目标的属性域，不能局限于转换中心认识的方块。
	properties, ok := java.BlockProperties("minecraft:copper_chest")
	if !ok || properties["facing"].Default != `"north"` {
		t.Fatalf("copper_chest 属性 = %v (%v)", properties, ok)
	}
	if _, exists := properties["waterlogged"]; exists {
		t.Fatal("单层属性域不应包含 waterlogged")
	}
}
