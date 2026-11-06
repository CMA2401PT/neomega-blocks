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

// 原理图本来就是 Java 数据；保留直接映射，不经过目标 Bedrock palette。
var schematicJava = sync.OnceValue(func() map[uint32]ir.IRBlock {
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
	for key, state := range states {
		idText, metaText, _ := strings.Cut(key, ":")
		id, err := strconv.ParseUint(idText, 10, 16)
		if err != nil {
			panic(err)
		}
		meta, err := strconv.ParseUint(metaText, 10, 16)
		if err != nil {
			panic(err)
		}
		result[uint32(id)<<16|uint32(meta)] = ir.IRBlock(state).Tidy()
	}
	return result
})

// SchematicToJava 把旧数字 ID/meta 直接转换为 Java 状态，未知值返回 false。
func SchematicToJava(id, meta uint16) (ir.IRBlock, bool) {
	state, ok := schematicJava()[uint32(id)<<16|uint32(meta)]
	return state, ok
}
