package v1

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"neomega-blocks/ir"
	specific_converts "neomega-blocks/presets/block-converts/specific-converts"
	"neomega-blocks/storage/num_lookup"
	"strings"
	"sync"
)

//go:embed 2_bedrock_fan_in.bin
var data2BedrockFanin []uint8

var bedrockFanin num_lookup.TwoLayerNumMap[uint32, uint32]
var bedrockFaninLoaded sync.Once

//go:embed 2_2_java_default_props.bin
var data2_2JavaDefaultPropsBytes []uint8
var javaDefaultProps map[string]ir.IRStates
var javaDefaultPropsLoaded sync.Once

//go:embed 2_3_java_fan_in.bin
var data2_3JavaFanin []uint8

var javaFanin num_lookup.TwoLayerNumMap[uint32, uint32]
var javaFaninLoaded sync.Once

func loadBedrockFanin() {
	// ensure load only once
	bedrockFaninLoaded.Do(func() {
		bedrockFanin = num_lookup.LoadTwoLayerNumMap[uint32, uint32](data2BedrockFanin)
	})
}

func loadJavaFanin() {
	// ensure load only once
	javaFaninLoaded.Do(func() {
		javaFanin = num_lookup.LoadTwoLayerNumMap[uint32, uint32](data2_3JavaFanin)
	})
}

// 尝试将一个基岩方块转到 当前版本的方块
func ConvertFromBedrock(bedrock ir.IRBlock) (ir.IRBlock, bool) {
	loadBedrockFanin()
	inHash := bedrock.Hash()
	// 尝试直接匹配
	outBlock, ok := BedrockHashToBlockStrict(inHash)
	if ok {
		return outBlock, true
	}
	// 尝试标准化
	cannonBlock, ok := BedrockCanon(bedrock)
	if ok {
		outBlock, ok := BedrockHashToBlockStrict(cannonBlock.Hash())
		if ok {
			return outBlock, true
		}
	}
	// 尝试从Fanin中读取
	outHash, ok := bedrockFanin.Get(inHash)
	if !ok {
		return "minecraft:unknown []", false
	}
	outBlock, ok = BedrockHashToBlockStrict(outHash)
	if ok {
		return outBlock, true
	}
	panic(fmt.Errorf("%v->%v(hash)-> block fail, should not happen as outHash exist", bedrock, outHash))
}

func loadjavaDefaultProps() {
	javaDefaultPropsLoaded.Do(func() {
		var defaults map[string]ir.IRStates
		if err := json.Unmarshal(data2_2JavaDefaultPropsBytes, &defaults); err != nil {
			panic(err)
		}
		javaDefaultProps = defaults
	})
}

// 通过方块名（含命名空间）获取默认属性。
func javaNameToDefaultProps(name string) (ir.IRStates, bool) {
	loadjavaDefaultProps()
	props, ok := javaDefaultProps[name]
	return props, ok
}

// JavaCanon 标准化 Java 的方块，根据记录删除不应该存在的key，并填补缺失的 key
// 如果失败则返回原值
func JavaCanon(blk ir.IRBlock) (ir.IRBlock, bool) {
	blk = blk.Tidy()
	props, ok := javaNameToDefaultProps(blk.NameWithNameSpace())
	if !ok {
		return blk, false
	}
	return ir.IRBlock(blk.NameWithNameSpace() + " " + string(blk.States().Canon(props))), true
}

func javaWallDirectConvert(inJava ir.IRBlock) (outBedrock ir.IRBlock, ok bool) {
	if !strings.HasSuffix(inJava.NameWithNameSpace(), "_wall") {
		return "", false
	}
	inJava, ok = JavaCanon(inJava)
	if !ok {
		return "", false
	}
	outStates := inJava.States().Update(specific_converts.JavaTobedrockWallRelplacer)
	bedrockBlock := ir.IRBlock(inJava.NameWithNameSpace() + " " + string(outStates))

	if HasBedrockBlockHash(bedrockBlock.Hash()) {
		return bedrockBlock, ok
	}
	return "", false
}

func javaHangingSignDirectConvert(inJava ir.IRBlock) (outBedrock ir.IRBlock, ok bool) {
	out, ok := specific_converts.JavaToBedrockHangingSign(inJava)
	if !ok {
		return "", false
	}
	if HasBedrockBlockHash(out.Hash()) {
		return out, ok
	}
	return "", false
}

// 尝试将一个 Java 方块转到 当前版本的方块
func ConvertFromJava(javaBlock ir.IRBlock) (ir.IRBlock, bool) {
	// 特殊情况直接硬编码转换方式
	if outBedrock, ok := javaWallDirectConvert(javaBlock); ok {
		return outBedrock, true
	}
	if outBedrock, ok := javaHangingSignDirectConvert(javaBlock); ok {
		return outBedrock, true
	}
	loadJavaFanin()
	// 尝试标准化
	cannonBlock, ok := JavaCanon(javaBlock)
	if ok {
		outHash, ok := javaFanin.Get(cannonBlock.Hash())
		if ok {
			outBlock, ok := BedrockHashToBlockStrict(outHash)
			if ok {
				return outBlock, true
			}
			panic(fmt.Errorf("%v->%v(hash)-> block fail, should not happen as outHash exist", javaBlock, outHash))
		}
	}
	// 尝试从Fanin中读取
	outHash, ok := javaFanin.Get(javaBlock.Hash())
	if ok {
		outBlock, ok := BedrockHashToBlockStrict(outHash)
		if ok {
			return outBlock, true
		}
		panic(fmt.Errorf("%v->%v(hash)-> block fail, should not happen as outHash exist", javaBlock, outHash))
	}
	return "minecraft:unknown []", false
}
