package blockconverts

import (
	"os"
	"strings"

	"neomega-blocks/ir"
)

func ReadJavaStates(path string) []ir.IRBlock {
	// 转换中心的文本保留上游属性顺序，读取时整理后才能与迁移键直接比较。
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var states []ir.IRBlock
	for line := range strings.Lines(string(data)) {
		states = append(states, ir.IRBlock(strings.TrimSpace(line)).Tidy())
	}
	return states
}

func ReadJavaMappings(path string) map[ir.IRBlock]ir.IRBlock {
	// 兼容表与迁移表共用源#目标格式，统一整理后供生成步骤连接。
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	mappings := make(map[ir.IRBlock]ir.IRBlock)
	for line := range strings.Lines(string(data)) {
		source, target, _ := strings.Cut(strings.TrimSpace(line), "#")
		mappings[ir.IRBlock(source).Tidy()] = ir.IRBlock(target).Tidy()
	}
	return mappings
}
