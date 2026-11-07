package blockconverts

import (
	"testing"

	"neomega-blocks/ir"
	converts "neomega-blocks/presets/block-converts/v1"
)

func TestBedrockPaletteSharedHashes(t *testing.T) {
	// 相同名称、状态与 NBT 类型应保留历史网络哈希；差异状态由正常转入规则解释。
	target := map[ir.IRBlock]uint32{}
	for hash, block := range converts.BedrockBlocks() {
		target[block.Tidy()] = hash
	}
	shared := 0
	for _, block := range ReadBlocks("block_palette.json") {
		hash, exists := target[block.IRBlock.Tidy()]
		if !exists {
			continue
		}
		shared++
		if hash != block.NetworkID {
			t.Fatalf("%s: hash=%d，历史网络哈希=%d", block.IRBlock, hash, block.NetworkID)
		}
	}
	t.Logf("同写法状态=%d，网络哈希全部一致", shared)
	for _, block := range []ir.IRBlock{
		`minecraft:stone []`,
		`minecraft:bed ["direction"=0,"head_piece_bit"=false,"occupied_bit"=false]`,
		`minecraft:observer ["minecraft:facing_direction"="east","powered_bit"=true]`,
	} {
		if _, exists := target[block.Tidy()]; !exists {
			t.Fatal(block)
		}
	}
}

func TestLegacyPaletteDataValues(t *testing.T) {
	// legacy aux 按历史 palette 返回原状态，即使该状态不在最新 Bedrock 全集中。
	for _, block := range ReadBlocks("block_palette.json") {
		if block.Data < 0 {
			continue
		}
		got, ok := converts.BedrockDataValueToBlock(block.IRBlock.NameWithNameSpace(), uint32(block.Data))
		if !ok || got != block.IRBlock.Tidy() {
			t.Fatalf("%s:%d → %s (%v)，期望 %s", block.IRBlock.ShortName(), block.Data, got, ok, block.IRBlock)
		}
	}
}
