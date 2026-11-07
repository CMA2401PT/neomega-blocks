package v26_3_test

import (
	"os"
	"testing"

	"neomega-blocks/catalog"
	"neomega-blocks/ir"
	legacy "neomega-blocks/presets/java/v1"
	java "neomega-blocks/presets/java/v26_3"
)

func TestEveryOfficialState(t *testing.T) {
	c, err := java.Open()
	if err != nil {
		t.Fatal(err)
	}
	if c.BlockCount() != 1286 || c.StateCount() != 35723 {
		t.Fatalf("counts = %d/%d", c.BlockCount(), c.StateCount())
	}
	if c.Metadata().DataVersion != 5023 || len(c.Metadata().Sources) != 5 {
		t.Fatal("missing provenance")
	}
	for state := range c.States() {
		byID, okID := c.ByID(state.ID)
		byText, okText := c.Lookup(state.Java)
		byHash, okHash := c.ByHash(state.Java.Hash())
		if !okID || !okText || !okHash || byID != state || byText != state || byHash != state {
			t.Fatalf("state lookup mismatch: %s", state.Java)
		}
		mapping, err := c.ToBedrock(state.Java)
		if err != nil || mapping.Block != state.Bedrock {
			t.Fatalf("mapping mismatch: %s %v", state.Java, err)
		}
		found := false
		for _, candidate := range c.JavaCandidates(mapping.Block) {
			if candidate == state {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("reverse candidates lost state %s", state.Java)
		}
	}
	for definition := range c.Blocks() {
		got, err := c.Normalize(ir.IRBlock(definition.Name + " []"))
		want, _ := c.ByID(definition.DefaultStateID)
		if err != nil || got != want.Java {
			t.Fatalf("official default mismatch: %s %v", definition.Name, err)
		}
	}
}

func TestNewBlocksWaterAndNBTTypes(t *testing.T) {
	c, err := java.Open()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"copper_chest", "sulfur", "cinnabar", "golden_dandelion", "test_instance_block", "poplar_planks", "red_wool_stairs", "white_concrete_slab"} {
		if _, ok := c.Definition(name); !ok {
			t.Errorf("new block missing: %s", name)
		}
	}
	slab, err := c.Normalize(`oak_slab []`)
	if err != nil || slab != `minecraft:oak_slab ["type"="bottom","waterlogged"="false"]` {
		t.Fatalf("slab default %s %v", slab, err)
	}
	dry, _ := c.ToBedrock(slab)
	wet, err := c.ToBedrock(`oak_slab ["waterlogged"="true"]`)
	if err != nil || !wet.Waterlogged || dry.Waterlogged || wet.Block != dry.Block {
		t.Fatalf("water flag lost: %#v %#v %v", dry, wet, err)
	}
	if len(c.JavaCandidates(wet.Block)) < 2 {
		t.Fatal("ambiguous inverse was collapsed")
	}
	if _, err := c.Normalize(`stone ["waterlogged"="true"]`); err == nil {
		t.Fatal("fabricated waterlogged stone accepted")
	}
	button, err := c.ToBedrock(`oak_button ["powered"="true"]`)
	if err != nil {
		t.Fatal(err)
	}
	pressed, _ := button.Block.States().Get("button_pressed_bit")
	facing, _ := button.Block.States().Get("facing_direction")
	if value, ok := pressed.(int8); !ok || value != 1 {
		t.Fatalf("byte type lost: %T %v", pressed, pressed)
	}
	if _, ok := facing.(int32); !ok {
		t.Fatalf("int type lost: %T %v", facing, facing)
	}
}

func TestLegacyPresetRetainsTarget(t *testing.T) {
	if legacy.JavaTargetVersion != "1.21.4" {
		t.Fatal("legacy target changed")
	}
	count := 0
	for range legacy.JavaFullStates() {
		count++
	}
	if count != 18254 {
		t.Fatalf("legacy states changed: %d", count)
	}
}

func BenchmarkLoad(b *testing.B) {
	data, err := os.ReadFile("catalog.json.gz")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := catalog.Load(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLookup(b *testing.B) {
	c, err := java.Open()
	if err != nil {
		b.Fatal(err)
	}
	block, err := c.Normalize(`copper_chest []`)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := c.Lookup(block); !ok {
			b.Fatal("state missing")
		}
	}
}

func BenchmarkByHash(b *testing.B) {
	c, err := java.Open()
	if err != nil {
		b.Fatal(err)
	}
	block, _ := c.Normalize(`copper_chest []`)
	hash := block.Hash()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := c.ByHash(hash); !ok {
			b.Fatal("state missing")
		}
	}
}
