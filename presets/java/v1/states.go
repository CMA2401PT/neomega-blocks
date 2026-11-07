package v1

import (
	_ "embed"
	"iter"
	"neomega-blocks/ir"
	"slices"
	"strings"
	"sync"
)

const JavaTargetVersion = "26.3"

// Java 目标版本的完整状态全集，由 third_party/gen_resource/block-converts/gen_fan_out_mapping.py 生成。
// 在模块根目录运行 sh third_party/gen_resource/block-converts/generate.sh，按依赖顺序更新资源。
//
//go:embed java_states.list
var dataJavaFullStates string

var loadStates = sync.OnceValues(func() ([]ir.IRBlock, map[uint32]ir.IRBlock) {
	states := make([]ir.IRBlock, 0)
	byHash := make(map[uint32]ir.IRBlock)
	for line := range strings.Lines(dataJavaFullStates) {
		block := ir.IRBlock(strings.TrimSpace(line)).Tidy()
		states = append(states, block)
		byHash[block.Hash()] = block
		// 含水是所有方块都可附带的状态，反查时保留它，供转换器和 fixer 传递。
		for _, value := range []string{"true", "false"} {
			properties := block.States().Set("waterlogged", value)
			variant := ir.IRBlock(block.NameWithNameSpace() + " " + string(properties)).Tidy()
			byHash[variant.Hash()] = variant
		}
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
	if !ok {
		block, ok = compatibilityByHash()[hash]
	}
	return block, ok
}

func JavaFullStates() iter.Seq[ir.IRBlock] {
	states, _ := loadStates()
	return slices.Values(states)
}
