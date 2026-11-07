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

// 目标 Bedrock 全集的网络哈希与方块，由 third_party/gen_resource/block-converts/step0_read_blocks_test.go
// 中的 TestStep0ReadBedrockStates 生成。
// 在模块根目录运行 sh third_party/gen_resource/block-converts/generate.sh，按依赖顺序更新资源。
//
//go:embed 0_bedrock_blocks.bin
var data0BedrockBlocksBytes []uint8

// Bedrock 默认属性，由 third_party/gen_resource/block-converts/step1_read_consistency_test.go
// 中的 TestReadDefaultsReadBlocksConsistency 生成。
// 在模块根目录运行 sh third_party/gen_resource/block-converts/generate.sh，按依赖顺序更新资源。
//
//go:embed 1_bedrock_default_props.bin
var data1BedrockDefaultPropsBytes []uint8

// 名称、data value 到网络哈希的索引，由 third_party/gen_resource/block-converts/step4_block_data_test.go
// 中的 TestStep4BuildBlockDataByName 生成。
// 在模块根目录运行 sh third_party/gen_resource/block-converts/generate.sh，按依赖顺序更新资源。
//
//go:embed 4_block_data.bin
var data4BlockDataBytes []uint8

// 旧物品 aux 指向的方块来自历史 palette（block_palette.json），与目标全集分开存放，
// 这样目标版本升级时旧 aux 仍能查到原来的状态。
// 由 third_party/gen_resource/block-converts/step4_block_data_test.go 中的 TestStep4BuildBlockDataByName 生成。
//
//go:embed 5_legacy_blocks.bin
var dataLegacyBlocks []byte

var legacyBlocks = sync.OnceValue(func() *blocks_name.BlockNameStorage[uint16] {
	return blocks_name.LoadBlockNameStorage[uint16](dataLegacyBlocks)
})

var bedrockNameStorage *blocks_name.BlockNameStorage[uint16]
var bedrockNameStorageLoaded sync.Once
var bedrockDefaultProps map[string]ir.IRStates
var bedrockDefaultPropsLoaded sync.Once
var bedrockDataValueToHash map[string][]uint32
var bedrockDataValueToHashLoaded sync.Once

//	func GetInstance() *Singleton {
//		once.Do(func() {
//			instance = &Singleton{}
//		})
//		return instance
//	}
func loadBedrockNameStorage() {
	// ensure load only once
	bedrockNameStorageLoaded.Do(func() {
		bedrockNameStorage = blocks_name.LoadBlockNameStorage[uint16](data0BedrockBlocksBytes)
	})
}

func loadBedrockDefaultProps() {
	bedrockDefaultPropsLoaded.Do(func() {
		var defaults map[string]ir.IRStates
		if err := json.Unmarshal(data1BedrockDefaultPropsBytes, &defaults); err != nil {
			panic(err)
		}
		bedrockDefaultProps = defaults
	})
}

func loadBedrockDataValueToHash() {
	bedrockDataValueToHashLoaded.Do(func() {
		loaded, err := loadStep4Storage(data4BlockDataBytes)
		if err != nil {
			panic(err)
		}
		bedrockDataValueToHash = loaded
	})
}

func HasBedrockBlockHash(hash uint32) bool {
	loadBedrockNameStorage()
	return bedrockNameStorage.Has(blocks_name.BlockHash(hash))
}

// BedrockHashToBlockStrict 只查询目标 Bedrock 全集中的状态。
func BedrockHashToBlockStrict(hash uint32) (ir.IRBlock, bool) {
	loadBedrockNameStorage()
	return bedrockNameStorage.Get(blocks_name.BlockHash(hash))
}

// BedrockHashToBlockData 查询历史 palette 中的名称与 data value。
func BedrockHashToBlockData(hash uint32) (string, uint16, bool) {
	return legacyBlocks().GetExtraValue(blocks_name.BlockHash(hash))
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
	loadBedrockDataValueToHash()
	arr, ok := bedrockDataValueToHash[name]
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
	return legacyBlocks().Get(blocks_name.BlockHash(hash))
}

// 通过方块名（含命名空间）获取默认属性。
func bedrockNameToDefaultProps(name string) (ir.IRStates, bool) {
	loadBedrockDefaultProps()
	props, ok := bedrockDefaultProps[name]
	return props, ok
}

// bedrockCanon 标准化 基岩版 的方块，根据记录删除不应该存在的key，并填补缺失的 key
// 如果失败则返回原值
func BedrockCanon(blk ir.IRBlock) (ir.IRBlock, bool) {
	blk = blk.Tidy()
	props, ok := bedrockNameToDefaultProps(blk.NameWithNameSpace())
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
	loadBedrockNameStorage()
	return func(yield func(uint32, ir.IRBlock) bool) {
		for hash := range bedrockNameStorage.NumLookup.All() {
			block, ok := bedrockNameStorage.Get(hash)
			if ok && !yield(uint32(hash), block) {
				return
			}
		}
	}
}
