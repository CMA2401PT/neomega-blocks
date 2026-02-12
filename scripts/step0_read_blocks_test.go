package scripts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"neomega-blocks/ir"
	blocks_name "neomega-blocks/storage/blocks-name"
)

// 从 NEMC 的 block_palette.json 中读取所有方块的状态，并保存到 0_bedrock_blocks.json
func TestStep0ReadBlocksFromPalette(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	palettePath := filepath.Join(filepath.Dir(cur), "block_palette.json")

	blocks := ReadBlocks(palettePath)
	if len(blocks) < 1000 {
		t.Fatalf("unexpected small block count: got %d", len(blocks))
	}

	for _, b := range blocks {
		if b.NetworkID != b.IRBlock.Hash() {
			t.Fatalf("hash mismatch %d!=%d", b.NetworkID, b.IRBlock.Hash())
		}
	}

	first := blocks[0]
	if first.BlockID != 987 {
		t.Fatalf("first.BlockID = %d, want 987", first.BlockID)
	}
	if first.NameHash != 35106087384923613 {
		t.Fatalf("first.NameHash = %d, want 35106087384923613", first.NameHash)
	}
	if first.NetworkID != 973836165 {
		t.Fatalf("first.NetworkID = %d, want 973836165", first.NetworkID)
	}
	if first.IRBlock != ir.IRBlock(`minecraft:cyan_terracotta []`) {
		t.Fatalf("first.IRBlock = %q, want %q", first.IRBlock, ir.IRBlock(`minecraft:cyan_terracotta []`))
	}

	foundBlueCandle := false
	for _, b := range blocks {
		if b.IRBlock == ir.IRBlock(`minecraft:blue_candle ["candles"=0,"lit"=false]`) {
			foundBlueCandle = true
			break
		}
	}
	if !foundBlueCandle {
		t.Fatalf("did not find expected IRBlock %q", ir.IRBlock(`minecraft:blue_candle ["candles"=0,"lit"=false]`))
	}

	storage := blocks_name.NewBlockNameStorage[uint16]()
	for _, b := range blocks {
		storage.Set(b.IRBlock)
	}

	for _, b := range blocks {
		got, ok := storage.Get(blocks_name.BlockHash(b.IRBlock.Hash()))
		if !ok {
			t.Fatalf("storage.Get(%q hash) found = false", b.IRBlock)
		}
		if got != b.IRBlock {
			t.Fatalf("storage.Get(%q hash) = %q, want %q", b.IRBlock, got, b.IRBlock)
		}
	}

	dumped := storage.Dump()
	loaded := blocks_name.LoadBlockNameStorage[uint16](dumped)

	for _, b := range blocks {
		got, ok := loaded.Get(blocks_name.BlockHash(b.IRBlock.Hash()))
		if !ok {
			t.Fatalf("loaded.Get(%q hash) found = false", b.IRBlock)
		}
		if got != b.IRBlock {
			t.Fatalf("loaded.Get(%q hash) = %q, want %q", b.IRBlock, got, b.IRBlock)
		}
	}
	fmt.Printf("%v MiB\n", float32(len(storage.Dump()))/(1<<20))
	os.WriteFile("../presets/v1/0_bedrock_blocks.bin", storage.Dump(), 0644)

	outNames := []uint8{}
	for _, b := range blocks {
		outNames = append(outNames, []uint8(b.IRBlock)...)
	}
	fmt.Printf("%v MiB\n", float32(len(outNames))/(1<<20))
	// os.WriteFile("dump_name.bin", outNames, 0644)
}
