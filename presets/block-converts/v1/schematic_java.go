package v1

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"neomega-blocks/ir"
	"strconv"
	"strings"
	"sync"
)

// 由 third_party/gen_resource/block-converts 的 TestSchematic 从 schematic.json 生成。
//
//go:embed 7_schematic_to_java.json.gz
var schematicJavaData []byte

// schematicStates 是「数字 ID:meta」到 Java 状态的映射，键由 schematicKey 生成。
// 原理图本来就是 Java 数据，所以直接查这张表，不经过目标 Bedrock palette，避免状态损失。
var schematicStates = sync.OnceValue(func() map[uint32]ir.IRBlock {
	reader, err := gzip.NewReader(bytes.NewReader(schematicJavaData))
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	var states map[string]string
	if err = json.NewDecoder(reader).Decode(&states); err != nil {
		panic(err)
	}
	result := make(map[uint32]ir.IRBlock, len(states))
	for entry, state := range states {
		idText, metaText, _ := strings.Cut(entry, ":")
		id, err := strconv.ParseUint(idText, 10, 16)
		if err != nil {
			panic(err)
		}
		meta, err := strconv.ParseUint(metaText, 10, 16)
		if err != nil {
			panic(err)
		}
		result[schematicKey(uint16(id), uint16(meta))] = ir.IRBlock(state).Tidy()
	}
	return result
})

// schematicKey 把数字 ID 和 meta 打包成映射的键：高 16 位是 ID，低 16 位是 meta。
func schematicKey(id, meta uint16) uint32 {
	return uint32(id)<<16 | uint32(meta)
}

// SchematicToJava 把旧原理图的数字 ID 与 meta 直接转成 Java 状态；表里没有时返回 false。
func SchematicToJava(id, meta uint16) (ir.IRBlock, bool) {
	state, ok := schematicStates()[schematicKey(id, meta)]
	return state, ok
}
