package v1

import (
	_ "embed"
	"iter"
	"neomega-blocks/ir"
	"slices"
	"strings"
	"sync"
)

const JavaTargetVersion = "1.21.4"

//go:embed fan_out_java_full.list
var dataJavaFullStates string

var loadStates = sync.OnceValues(func() ([]ir.IRBlock, map[uint32]ir.IRBlock) {
	states := make([]ir.IRBlock, 0)
	byHash := make(map[uint32]ir.IRBlock)
	for line := range strings.Lines(dataJavaFullStates) {
		block := ir.IRBlock(strings.TrimSpace(line)).Tidy()
		states = append(states, block)
		byHash[block.Hash()] = block
	}
	return states, byHash
})

func BlockToHash(block ir.IRBlock) uint32 {
	// 先规范属性顺序，让同一完整状态的哈希与输入顺序无关。
	return block.Tidy().Hash()
}

func HashToBlock(hash uint32) (ir.IRBlock, bool) {
	_, byHash := loadStates()
	block, ok := byHash[hash]
	return block, ok
}

func JavaFullStates() iter.Seq[ir.IRBlock] {
	states, _ := loadStates()
	return slices.Values(states)
}
