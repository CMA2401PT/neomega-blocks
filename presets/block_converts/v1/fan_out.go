package v1

import (
	_ "embed"
	"neomega-blocks/ir"
	"neomega-blocks/specific_converts"
	blocks_name "neomega-blocks/storage/blocks-name"
	"strings"
	"sync"
)

//go:embed 3_to_java.bin
var data3ToJavaBytes []uint8

var toJavaNameStorage *blocks_name.BlockNameStorage[uint16]
var toJavaNameStorageLoaded sync.Once

func loadToJavaNameStorage() {
	// ensure load only once
	toJavaNameStorageLoaded.Do(func() {
		toJavaNameStorage = blocks_name.LoadBlockNameStorage[uint16](data3ToJavaBytes)
	})
}

func bedrockWallDirectConvert(inBedrock ir.IRBlock) (outJava ir.IRBlock, ok bool) {
	if !strings.HasSuffix(inBedrock.NameWithNameSpace(), "_wall") {
		return "", false
	}
	inBedrock, ok = NEMCBedrockCanon(inBedrock)
	if !ok {
		return "", false
	}
	outStates := inBedrock.States().Update(specific_converts.BedrockToJavaWallRelplacer)
	return ir.IRBlock(inBedrock.NameWithNameSpace() + " " + string(outStates)), true
}
func bedrockHangingSignDirectConvert(inBedrock ir.IRBlock) (outJava ir.IRBlock, ok bool) {
	if outJava, ok := specific_converts.BedrockToJavaHangingSign(inBedrock); ok {
		return outJava, ok
	}
	return inBedrock, false
}

// 尝试将一个当前的 block 转为 java block
func ConvertToJava(bedrock ir.IRBlock) (javaBlock ir.IRBlock, ok bool) {
	// 特殊情况直接硬编码转换方式
	if javaBlock, ok = bedrockWallDirectConvert(bedrock); ok {
		return javaBlock, true
	}
	if javaBlock, ok = bedrockHangingSignDirectConvert(bedrock); ok {
		return javaBlock, true
	}
	bedrock, ok = NEMCBedrockCanon(bedrock)
	if !ok {
		return "", false
	}
	loadToJavaNameStorage()
	return toJavaNameStorage.Get(blocks_name.BlockHash(bedrock.Hash()))
}
