package solidtable

import (
	"bytes"
	_ "embed"
	"sync"

	"neomega-blocks/geometry/solidtable"
)

// 由 third_party/gen_resource/geometry/solid_block_table 与几何表一起生成。
// 在模块根目录运行 go generate ./presets/java/v1/geometry。
//
//go:embed geometry.nvsolid
var bundled []byte

// Bundled 返回随包分发的实心表；多个调用方共享同一份只读解码结果。
var Bundled = sync.OnceValues(func() (solidtable.Table, error) {
	return solidtable.Decode(bytes.NewReader(bundled))
})
