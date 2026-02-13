package scripts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

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
	os.WriteFile("../v1/0_bedrock_blocks.bin", storage.Dump(), 0644)

	outNames := []uint8{}
	for _, b := range blocks {
		outNames = append(outNames, []uint8(b.IRBlock)...)
	}
	fmt.Printf("%v MiB\n", float32(len(outNames))/(1<<20))
	// os.WriteFile("dump_name.bin", outNames, 0644)
}
