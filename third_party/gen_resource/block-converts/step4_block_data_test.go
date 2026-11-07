//go:build resourcegen

package blockconverts

import (
	"fmt"
	blocks_name "neomega-blocks/storage/blocks-name"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestStep4BuildBlockDataByName(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	palettePath := filepath.Join(filepath.Dir(cur), "block_palette.json")

	blocks := ReadBlocks(palettePath)
	if len(blocks) < 1000 {
		t.Fatalf("unexpected small block count: got %d", len(blocks))
	}
	legacy := blocks_name.NewBlockNameStorage[uint16]()
	for _, block := range blocks {
		legacy.Set(block.IRBlock)
		legacy.SetExtraValue(blocks_name.BlockHash(block.NetworkID), uint16(block.Data))
	}
	if err := os.WriteFile("../../../presets/block-converts/v1/5_legacy_blocks.bin", legacy.Dump(), 0o644); err != nil {
		t.Fatal(err)
	}

	storage := BuildStep4Storage(blocks)
	emptyCount := 0
	nonEmptyCount := 0
	for _, vs := range storage {
		for _, v := range vs {
			if v == 0 {
				emptyCount += 1
			} else {
				nonEmptyCount += 1
			}
		}
	}
	fmt.Println(emptyCount, nonEmptyCount)

	if err := os.WriteFile("../../../presets/block-converts/v1/4_block_data.bin", storage.Dump(), 0o644); err != nil {
		t.Fatalf("write 4_block_data.bin: %v", err)
	}
}

func TestStep4StorageDumpLoad(t *testing.T) {
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	palettePath := filepath.Join(filepath.Dir(cur), "block_palette.json")
	blocks := ReadBlocks(palettePath)

	storage := BuildStep4Storage(blocks)
	data := storage.Dump()
	loaded, err := LoadStep4Storage(data)
	if err != nil {
		t.Fatalf("LoadStep4Storage failed: %v", err)
	}

	if len(loaded) != len(storage) {
		t.Fatalf("loaded len=%d, want=%d", len(loaded), len(storage))
	}
	for name, want := range storage {
		got, ok := loaded[name]
		if !ok {
			t.Fatalf("missing key after load: %s", name)
		}
		if len(got) != len(want) {
			t.Fatalf("name=%s len(got)=%d len(want)=%d", name, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("name=%s idx=%d got=%d want=%d", name, i, got[i], want[i])
			}
		}
	}
}
