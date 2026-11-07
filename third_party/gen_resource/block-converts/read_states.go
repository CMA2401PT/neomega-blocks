package blockconverts

import (
	"os"
	"strings"

	"neomega-blocks/ir"
)

func ReadStates(path string) []ir.IRBlock {
	// 上游文本读取后统一属性顺序，供原理图和往返分析比较。
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
