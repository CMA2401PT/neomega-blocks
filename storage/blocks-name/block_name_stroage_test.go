package blocks_name

import (
	"fmt"
	"testing"

	"neomega-blocks/ir"
)

func TestBlockNameStorageSetAndGetByHash(t *testing.T) {
	blocks := generateLargeBlocks(120, 12)
	s := NewBlockNameStorage[uint32]()
	wantByHash := make(map[BlockHash]ir.IRBlock, len(blocks))

	for _, blk := range blocks {
		h := BlockHash(blk.Hash())
		if prev, ok := wantByHash[h]; ok && prev != blk {
			// hash 冲突会覆盖；这里跳过冲突样本，避免非确定性断言。
			continue
		}
		wantByHash[h] = blk
		s.Set(blk)
	}

	for h, want := range wantByHash {
		got, ok := s.Get(h)
		if !ok {
			t.Fatalf("Get(hash=%d) found = false", h)
		}
		if got != want {
			t.Fatalf("Get(hash=%d) = %q, want %q", h, got, want)
		}
	}

	if got, ok := s.Get(BlockHash(0)); ok || got != "" {
		t.Fatalf("Get(missing hash) = (%q, %v), want (\"\", false)", got, ok)
	}
}

func TestBlockNameStorageDumpLoadLargeScale(t *testing.T) {
	blocks := generateLargeBlocks(220, 16)
	s := NewBlockNameStorage[uint32]()
	wantByHash := make(map[BlockHash]ir.IRBlock, len(blocks))

	for _, blk := range blocks {
		h := BlockHash(blk.Hash())
		if prev, ok := wantByHash[h]; ok && prev != blk {
			continue
		}
		wantByHash[h] = blk
		s.Set(blk)
	}

	dumped := s.Dump()
	loaded := LoadBlockNameStorage[uint32](dumped)

	for h, want := range wantByHash {
		got, ok := loaded.Get(h)
		if !ok {
			t.Fatalf("loaded.Get(hash=%d) found = false", h)
		}
		if got != want {
			t.Fatalf("loaded.Get(hash=%d) = %q, want %q", h, got, want)
		}
	}

	if got, ok := loaded.Get(BlockHash(0)); ok || got != "" {
		t.Fatalf("loaded.Get(missing hash) = (%q, %v), want (\"\", false)", got, ok)
	}
}

func generateLargeBlocks(blockNameCount, statesPerName int) []ir.IRBlock {
	out := make([]ir.IRBlock, 0, blockNameCount*statesPerName)
	for i := 0; i < blockNameCount; i++ {
		name := fmt.Sprintf("minecraft:test_block_%03d", i)
		for j := 0; j < statesPerName; j++ {
			blk := ir.IRBlock(fmt.Sprintf(
				`%s ["a"=%d,"b"=%d,"c"="%03d_%03d","d"=%v]`,
				name,
				j,
				(i*statesPerName+j)%97,
				i, j,
				j%2 == 0,
			))
			out = append(out, blk)
		}
	}
	return out
}
