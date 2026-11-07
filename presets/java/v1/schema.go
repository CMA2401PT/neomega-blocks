package v1

import (
	_ "embed"
	"strings"
	"sync"
)

type Property struct {
	// 保留 SNBT 文本的引号，以便直接拼入 IRStates。
	Default string
	Values  []string
}

// Java 目标版本的属性域和默认值，由 third_party/gen_resource/java-states 的 TestGenerateJavaStates 生成。
// 在模块根目录运行 sh third_party/gen_resource/block-converts/generate.sh，按依赖顺序更新资源。
//
//go:embed java_block_schema.list
var dataBlockSchema string

var loadProperties = sync.OnceValue(func() map[string]map[string]Property {
	blocks := make(map[string]map[string]Property)
	for line := range strings.Lines(dataBlockSchema) {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		properties := make(map[string]Property, len(fields)-1)
		for _, field := range fields[1:] {
			name, spec, _ := strings.Cut(field, "=")
			defaultValue, domain, _ := strings.Cut(spec, "|")
			properties[name] = Property{
				Default: defaultValue,
				Values:  strings.Split(domain, ","),
			}
		}
		blocks[fields[0]] = properties
	}
	return blocks
})

func BlockProperties(name string) (map[string]Property, bool) {
	// 返回的属性表和取值切片共享缓存，调用方只读使用。
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	properties, ok := loadProperties()[name]
	return properties, ok
}
