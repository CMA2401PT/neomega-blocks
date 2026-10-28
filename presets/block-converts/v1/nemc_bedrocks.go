package v1

import (
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"iter"
	"neomega-blocks/ir"
	blocks_name "neomega-blocks/storage/blocks-name"
	"strings"
	"sync"
)

//go:embed 0_bedrock_blocks.bin
var data0BedrockBlocksBytes []uint8

//go:embed 1_bedrock_default_props.bin
var data1BedrockDefaultPropsBytes []uint8

//go:embed 4_block_data.bin
var data4BlockDataBytes []uint8

var nemcBedrockNameStorage *blocks_name.BlockNameStorage[uint16]
var nemcBedrockNameStorageLoaded sync.Once
var nemcBedrockDefaultProps map[string]ir.IRStates
var nemcBedrockDefaultPropsLoaded sync.Once
var nemcBedrockDataValueToHash map[string][]uint32
var nemcBedrockDataValueToHashLoaded sync.Once

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

func loadNemcBedrockDataValueToHash() {
	nemcBedrockDataValueToHashLoaded.Do(func() {
		loaded, err := loadStep4Storage(data4BlockDataBytes)
		if err != nil {
			panic(err)
		}
		nemcBedrockDataValueToHash = loaded
	})
}

func HasBedrockBlockHash(hash uint32) bool {
	loadNemcBedrockNameStorage()
	return nemcBedrockNameStorage.Has(blocks_name.BlockHash(hash))
}

// 从 NEMC 的 hash 映射到 方块，必须保证 hash 在 NEMC 的表中存在
func BedrockHashToBlockStrict(hash uint32) (ir.IRBlock, bool) {
	loadNemcBedrockNameStorage()
	return nemcBedrockNameStorage.Get(blocks_name.BlockHash(hash))
}

// 从 NEMC 的 hash 映射到 数据值
func BedrockHashToBlockData(hash uint32) (string, uint16, bool) {
	loadNemcBedrockNameStorage()
	return nemcBedrockNameStorage.GetExtraValue(blocks_name.BlockHash(hash))
}

// BedrockDataValueToHash 将 legacy data value 映射到 hash。
// name 若缺少 namespace 会自动加上 "minecraft:" 前缀。
// 当 value 越界或该位置 hash 为 0 时，返回 (0,false)。
func BedrockDataValueToHash(name string, value uint32) (hash uint32, ok bool) {
	if name == "unknown" || name == "minecraft:unknown" {
		return ^uint32(1), true
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	loadNemcBedrockDataValueToHash()
	arr, ok := nemcBedrockDataValueToHash[name]
	if !ok {
		return 0, false
	}
	if int(value) >= len(arr) {
		return 0, false
	}
	hash = arr[value]
	if hash == 0 {
		return 0, false
	}
	return hash, true
}

// BedrockDataValueToBlock 将 legacy data value 映射到 IRBlock
// name 若缺少 namespace 会自动加上 "minecraft:" 前缀。
// 当 value 越界或该位置 hash 为 0 时，返回 minecraft:unknown [] false
func BedrockDataValueToBlock(name string, value uint32) (block ir.IRBlock, ok bool) {
	if name == "unknown" || name == "minecraft:unknown" {
		return ir.IRBlock("minecraft:unknown []"), true
	}
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	hash, ok := BedrockDataValueToHash(name, value)
	if !ok {
		return ir.IRBlock("minecraft:unknown []"), false
	}
	return BedrockHashToBlockStrict(hash)
}

// 通过方块名（含命名空间）获取默认属性。
func nemcBedrockNameToDefaultProps(name string) (ir.IRStates, bool) {
	loadNemcBedrockDefaultProps()
	props, ok := nemcBedrockDefaultProps[name]
	return props, ok
}

// bedrockCanon 标准化 基岩版 的方块，根据记录删除不应该存在的key，并填补缺失的 key
// 如果失败则返回原值
func BedrockCanon(blk ir.IRBlock) (ir.IRBlock, bool) {
	blk = blk.Tidy()
	props, ok := nemcBedrockNameToDefaultProps(blk.NameWithNameSpace())
	if !ok {
		return blk, false
	}
	return ir.IRBlock(blk.NameWithNameSpace() + " " + string(blk.States().Canon(props))), true
}

func loadStep4Storage(data []byte) (map[string][]uint32, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid step4 storage: too short")
	}
	if string(data[:4]) != "BDS1" {
		return nil, fmt.Errorf("invalid step4 storage: bad magic")
	}
	off := 4
	count := int(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4

	out := make(map[string][]uint32, count)
	for i := 0; i < count; i++ {
		if off+4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: missing key length")
		}
		keyLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		if keyLen < 0 || off+keyLen > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: bad key length")
		}
		key := string(data[off : off+keyLen])
		off += keyLen

		if off+4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: missing array length")
		}
		arrLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		if arrLen < 0 || off+arrLen*4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: bad array length")
		}
		arr := make([]uint32, arrLen)
		for j := 0; j < arrLen; j++ {
			arr[j] = binary.LittleEndian.Uint32(data[off : off+4])
			off += 4
		}
		out[key] = arr
	}
	if off != len(data) {
		return nil, fmt.Errorf("invalid step4 storage: trailing bytes")
	}
	return out, nil
}

// BedrockBlocks iterates the preset's actual ID space without exposing its
// compact storage. IDs are network hashes; no new viewer-specific IDs are made.
func BedrockBlocks() iter.Seq2[uint32, ir.IRBlock] {
	loadNemcBedrockNameStorage()
	return func(yield func(uint32, ir.IRBlock) bool) {
		for hash := range nemcBedrockNameStorage.NumLookup.All() {
			block, ok := nemcBedrockNameStorage.Get(hash)
			if ok && !yield(uint32(hash), block) {
				return
			}
		}
	}
}
