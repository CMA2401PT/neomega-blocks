package v1

import (
	_ "embed"
	"encoding/json"
	"neomega-blocks/ir"
	blocks_name "neomega-blocks/storage/blocks-name"
	"sync"
)

//go:embed 0_bedrock_blocks.bin
var data0BedrockBlocksBytes []uint8

//go:embed 1_bedrock_default_props.bin
var data1BedrockDefaultPropsBytes []uint8

var nemcBedrockNameStorage *blocks_name.BlockNameStorage[uint16]
var nemcBedrockNameStorageLoaded sync.Once
var nemcBedrockDefaultProps map[string]ir.IRStates
var nemcBedrockDefaultPropsLoaded sync.Once

//	func GetInstance() *Singleton {
//		once.Do(func() {
//			instance = &Singleton{}
//		})
//		return instance
//	}
func loadNemcBedrockNameStorage() {
	// ensure load only once
	nemcBedrockNameStorageLoaded.Do(func() {
		nemcBedrockNameStorage = blocks_name.LoadBlockNameStorage[uint16](data0BedrockBlocksBytes)
	})
}

func loadNemcBedrockDefaultProps() {
	nemcBedrockDefaultPropsLoaded.Do(func() {
		var defaults map[string]ir.IRStates
		if err := json.Unmarshal(data1BedrockDefaultPropsBytes, &defaults); err != nil {
			panic(err)
		}
		nemcBedrockDefaultProps = defaults
	})
}

func HasNEMCBlockHash(hash uint32) bool {
	loadNemcBedrockNameStorage()
	return nemcBedrockNameStorage.Has(blocks_name.BlockHash(hash))
}

// 从 NEMC 的 hash 映射到 方块，必须保证 hash 在 NEMC 的表中存在
func NEMCBedrockHashToBlockStrict(hash uint32) (ir.IRBlock, bool) {
	loadNemcBedrockNameStorage()
	return nemcBedrockNameStorage.Get(blocks_name.BlockHash(hash))
}

// 通过方块名（含命名空间）获取默认属性。
func nemcBedrockNameToDefaultProps(name string) (ir.IRStates, bool) {
	loadNemcBedrockDefaultProps()
	props, ok := nemcBedrockDefaultProps[name]
	return props, ok
}

// bedrockCanon 标准化 基岩版 的方块，根据记录删除不应该存在的key，并填补缺失的 key
// 如果失败则返回原值
func NEMCBedrockCanon(blk ir.IRBlock) (ir.IRBlock, bool) {
	blk = blk.Tidy()
	props, ok := nemcBedrockNameToDefaultProps(blk.NameWithNameSpace())
	if !ok {
		return blk, false
	}
	return ir.IRBlock(blk.NameWithNameSpace() + " " + string(blk.States().Canon(props))), true
}
