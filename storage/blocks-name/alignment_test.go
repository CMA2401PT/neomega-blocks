package blocks_name

import (
	"neomega-blocks/ir"
	"testing"
	"unsafe"
)

func TestLoadUnalignedBlockStorage(t *testing.T) {
	original := NewBlockNameStorage[uint16]()
	blocks := []ir.IRBlock{`minecraft:air []`, `minecraft:bed ["direction"=1]`, `minecraft:bed ["direction"=2]`}
	for _, block := range blocks {
		original.Set(block)
	}
	data := original.Dump()
	backing := make([]uint32, (len(data)+4)/4)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&backing[0])), len(backing)*4)
	unaligned := raw[1 : 1+len(data)]
	copy(unaligned, data)
	loaded := LoadBlockNameStorage[uint16](unaligned)
	for _, block := range blocks {
		if got, ok := loaded.Get(BlockHash(block.Hash())); !ok || got != block {
			t.Fatalf("got %s,%v; want %s", got, ok, block)
		}
	}
}
