package v1

import (
	_ "embed"
	"bufio"
	"iter"
	"neomega-blocks/ir"
	"strings"
)

// Java 1.21.4 所有方块的完整状态全集，每行一个字段齐全的 blockstate。
// 由 scripts/gen_fan_out_mapping.py 对 PyMCTranslate 的属性取值域做完整笛卡尔积生成。
// 它区分 Bedrock 表达不了的状态（如北连的栅栏），供渲染侧建 hash→状态 的索引。
//
//go:embed fan_out_java_full.list
var dataJavaFullStates string

// JavaFullStates 遍历完整状态全集，每个状态已 Tidy 为字典序规范形式。
// 调用方对它算 IRBlock.Hash() 得到的值，与同样 Tidy 后的输入哈希一致。
func JavaFullStates() iter.Seq[ir.IRBlock] {
	return func(yield func(ir.IRBlock) bool) {
		scanner := bufio.NewScanner(strings.NewReader(dataJavaFullStates))
		scanner.Buffer(make([]byte, 1<<20), 1<<20)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !yield(ir.IRBlock(line).Tidy()) {
				return
			}
		}
	}
}
