package blocks_name

import (
	"encoding/binary"
	"fmt"
	"neomega-blocks/ir"
	compact_radix "neomega-blocks/storage/compact-radix"
	"neomega-blocks/storage/num_lookup"
)

type DumpableLookup[EP ~uint8 | ~uint16 | ~uint32] interface {
	compact_radix.EdgeLookup[EP]
	Dump() []uint8
	Load(data []uint8)
}

type RebuildIdx uint32
type BlockHash uint32

func (idx RebuildIdx) BlockNameIdx() uint16 {
	return uint16(idx >> 16)
}

func (idx RebuildIdx) StatesIdx() uint16 {
	return uint16(idx)
}

type blockSet[EP ~uint8 | ~uint16 | ~uint32] struct {
	*compact_radix.CompactRadix[EP, uint16, *compact_radix.ByteEdgeLookup[EP]]
	idx uint16
}

// BlockNameStorage 目的是为了通过 BlockHash 重构出 IRBlock
type BlockNameStorage[EP ~uint8 | ~uint16 | ~uint32] struct {
	Lookup *compact_radix.ByteEdgeLookup[EP]
	// Blocks 中的 CompactRadix 共享这个 Lookup
	Blocks map[string]blockSet[EP]
	// BlocksByIdx 用于从 RebuildIdx 反向构建 IRBlock
	BlocksByIdx map[uint16]*compact_radix.CompactRadix[EP, uint16, *compact_radix.ByteEdgeLookup[EP]]
	// 用于将 BlockHash 转换为 RebuildIdx
	NumLookup *num_lookup.TwoLayerNumMap[BlockHash, RebuildIdx]
	// debugNumLookup map[BlockHash]RebuildIdx
}

func NewBlockNameStorage[EP ~uint8 | ~uint16 | ~uint32]() *BlockNameStorage[EP] {
	lookup := compact_radix.NewByteEdgeLookup[EP]()
	return &BlockNameStorage[EP]{
		Lookup:      lookup,
		Blocks:      map[string]blockSet[EP]{},
		BlocksByIdx: map[uint16]*compact_radix.CompactRadix[EP, uint16, *compact_radix.ByteEdgeLookup[EP]]{},
		NumLookup:   num_lookup.NewTwoLayerNumMap[BlockHash, RebuildIdx](),
		// debugNumLookup: make(map[BlockHash]RebuildIdx),
	}
}

func (s *BlockNameStorage[EP]) Set(blk ir.IRBlock) {
	name := blk.NameWithNameSpace()
	states := blk.States().Tidy()
	set, ok := s.Blocks[name]
	if !ok {
		if len(s.Blocks) > int(^uint16(0)) {
			panic("block name count overflow")
		}
		idx := uint16(len(s.Blocks))
		set = blockSet[EP]{
			CompactRadix: compact_radix.NewCompactRadixWithLookup[EP, uint16, *compact_radix.ByteEdgeLookup[EP]](s.Lookup),
			idx:          idx,
		}
		s.Blocks[name] = set
		s.BlocksByIdx[idx] = set.CompactRadix
	}
	stateIdx := set.CompactRadix.Set([]uint8(states), 0)
	rebuildIdx := RebuildIdx(uint32(set.idx)<<16) | RebuildIdx(stateIdx)
	s.NumLookup.Set(BlockHash(blk.Hash()), rebuildIdx)
	// s.debugNumLookup[BlockHash(blk.Hash())] = rebuildIdx
}

func (s *BlockNameStorage[EP]) Get(hash BlockHash) (value ir.IRBlock, found bool) {
	rebuildIdx, ok := s.NumLookup.Get(hash)
	if !ok {
		return "", false
	}
	// if expect, ok := s.debugNumLookup[hash]; ok && expect != rebuildIdx {
	// 	panic("not equal")
	// }
	name, ok := s.blockNameByIdx(rebuildIdx.BlockNameIdx())
	if !ok {
		return "", false
	}
	set := s.BlocksByIdx[rebuildIdx.BlockNameIdx()]
	if set == nil {
		return "", false
	}
	statePath := set.Rebuild(rebuildIdx.StatesIdx())
	return ir.IRBlock(name + " " + string(statePath)), true
}

func (s *BlockNameStorage[EP]) Dump() []uint8 {
	// 将 s 中的信息导出用于以后的重构
	// s.BlocksByIdx 理论上不需要导出，因为可以从 s.Blocks 中获得
	// s.Lookup 也只需导出一次，因为所有都在共享一个 Lookup
	lookupData := s.Lookup.Dump()
	numLookupData := s.NumLookup.Dump()

	total := 4 + 4 + len(lookupData)
	total += (4 - (total % 4)) % 4
	total += 4 + len(numLookupData) + 2
	nodesByIdx := map[uint16][]byte{}
	namesByIdx := map[uint16]string{}
	for name, set := range s.Blocks {
		namesByIdx[set.idx] = name
		nodesByIdx[set.idx] = set.DumpNodes()
		// fmt.Println(name, len(nodesByIdx[set.idx]))
	}
	if len(namesByIdx) > int(^uint16(0)) {
		panic("block name count overflow")
	}
	for idx := uint16(0); idx < uint16(len(namesByIdx)); idx++ {
		name, ok := namesByIdx[idx]
		if !ok {
			panic(fmt.Errorf("missing block-name index %d", idx))
		}
		nodes, ok := nodesByIdx[idx]
		if !ok {
			panic(fmt.Errorf("missing block nodes for index %d", idx))
		}
		total += 2 + len(name) + 4
		pad := (4 - (total % 4)) % 4
		total += pad + len(nodes)
	}

	out := make([]byte, 0, total)
	out = append(out, 'B', 'N', 'S', '2')
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(len(lookupData)))
	out = append(out, buf4...)
	out = append(out, lookupData...)
	lookupPad := (4 - (len(out) % 4)) % 4
	for i := 0; i < lookupPad; i++ {
		out = append(out, 0)
	}
	binary.LittleEndian.PutUint32(buf4, uint32(len(numLookupData)))
	out = append(out, buf4...)
	out = append(out, numLookupData...)
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(len(namesByIdx)))
	out = append(out, buf2...)

	for idx := uint16(0); idx < uint16(len(namesByIdx)); idx++ {
		name := namesByIdx[idx]
		nodes := nodesByIdx[idx]
		binary.LittleEndian.PutUint16(buf2, uint16(len(name)))
		out = append(out, buf2...)
		out = append(out, []byte(name)...)
		binary.LittleEndian.PutUint32(buf4, uint32(len(nodes)))
		out = append(out, buf4...)
		pad := (4 - (len(out) % 4)) % 4
		for i := 0; i < pad; i++ {
			out = append(out, 0)
		}
		out = append(out, nodes...)
	}
	return out
}

func LoadBlockNameStorage[EP ~uint8 | ~uint16 | ~uint32](data []uint8) *BlockNameStorage[EP] {
	// 从 dump 的结果中还原，使用 CompactRadix.LoadNodes 以尽量避免分配
	// 保证 data 使用完后不会被释放
	if len(data) < 14 {
		panic("invalid block-name storage dump: too short")
	}
	if string(data[:4]) != "BNS2" {
		panic("invalid block-name storage dump: bad magic")
	}
	off := 4

	lookupLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4
	if lookupLen < 0 || off+lookupLen > len(data) {
		panic("invalid block-name storage dump: bad lookup size")
	}

	s := NewBlockNameStorage[EP]()
	s.Lookup.Load(data[off : off+lookupLen])
	off += lookupLen
	lookupPad := (4 - (off % 4)) % 4
	if off+lookupPad > len(data) {
		panic("invalid block-name storage dump: bad lookup alignment padding")
	}
	off += lookupPad

	if off+4 > len(data) {
		panic("invalid block-name storage dump: missing num-lookup size")
	}
	numLookupLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4
	if numLookupLen < 0 || off+numLookupLen > len(data) {
		panic("invalid block-name storage dump: bad num-lookup size")
	}
	loadedNumLookup := num_lookup.LoadTwoLayerNumMap[BlockHash, RebuildIdx](data[off : off+numLookupLen])
	s.NumLookup = &loadedNumLookup
	off += numLookupLen

	if off+2 > len(data) {
		panic("invalid block-name storage dump: missing block count")
	}
	blockCount := int(binary.LittleEndian.Uint16(data[off : off+2]))
	off += 2

	for i := 0; i < blockCount; i++ {
		if off+2 > len(data) {
			panic("invalid block-name storage dump: missing block name length")
		}
		nameLen := int(binary.LittleEndian.Uint16(data[off : off+2]))
		off += 2
		if nameLen < 0 || off+nameLen > len(data) {
			panic("invalid block-name storage dump: bad block name length")
		}
		name := string(data[off : off+nameLen])
		off += nameLen

		if off+4 > len(data) {
			panic("invalid block-name storage dump: missing node length")
		}
		nodesLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		pad := (4 - (off % 4)) % 4
		if off+pad > len(data) {
			panic("invalid block-name storage dump: bad node alignment padding")
		}
		off += pad
		if nodesLen < 0 || off+nodesLen > len(data) {
			panic("invalid block-name storage dump: bad node length")
		}

		r := compact_radix.NewCompactRadixWithLookup[EP, uint16, *compact_radix.ByteEdgeLookup[EP]](s.Lookup)
		r.LoadNodes(data[off : off+nodesLen])
		off += nodesLen

		idx := uint16(i)
		set := blockSet[EP]{CompactRadix: r, idx: idx}
		s.Blocks[name] = set
		s.BlocksByIdx[idx] = r
	}

	if off != len(data) {
		panic("invalid block-name storage dump: trailing bytes")
	}
	return s
}

func (s *BlockNameStorage[EP]) blockNameByIdx(idx uint16) (string, bool) {
	for name, set := range s.Blocks {
		if set.idx == idx {
			return name, true
		}
	}
	return "", false
}
