package v1

import (
	"neomega-blocks/ir"
	"testing"
)

func TestCompatibilityUsesNativeTargetsAndPreservesSourceIdentity(t *testing.T) {
	native := map[ir.IRBlock]bool{}
	seen := map[uint32]ir.IRBlock{}
	for block := range JavaFullStates() {
		native[block] = true
		seen[block.Hash()] = block
	}
	for mapping := range JavaCompatibility() {
		if !native[mapping.Target] || native[mapping.Source] || !mapping.Source.IsTidy() {
			t.Fatalf("invalid compatibility pair: %+v", mapping)
		}
		if previous, exists := seen[mapping.Source.Hash()]; exists {
			t.Fatalf("duplicate source: %s / %s", previous, mapping.Source)
		}
		seen[mapping.Source.Hash()] = mapping.Source
		if _, exists := mapping.Source.States().Get("waterlogged"); exists {
			t.Fatal("compatibility source contains waterlogged")
		}
		if got, ok := NormalizeSingleLayer(mapping.Source); !ok || got != mapping.Target {
			t.Fatal(mapping)
		}
	}
	for _, source := range []ir.IRBlock{"minecraft:grass []", `minecraft:grass ["waterlogged"="true"]`} {
		got, ok := NormalizeSingleLayer(source)
		if !ok || got != "minecraft:short_grass []" {
			t.Fatalf("newest grass interpretation lost: %s", got)
		}
		original, ok := HashToBlock(source.Hash())
		if !ok || original != source {
			t.Fatalf("reverse lookup lost source state: %s", original)
		}
	}
	if _, ok := NormalizeSingleLayer("example:unknown []"); ok {
		t.Fatal("unknown input accepted")
	}
}
