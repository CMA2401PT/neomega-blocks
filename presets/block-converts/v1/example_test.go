package v1_test

import (
	"fmt"
	"neomega-blocks/ir"
	blocks "neomega-blocks/presets/block-converts/v1"
)

func ExampleConvertFromJava() {
	// 1. Java 属性中的布尔值是字符串；Tidy 统一文本，转换入口补齐缺省属性。
	var input ir.IRBlock = ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`).Tidy()
	var bedrock ir.IRBlock
	var ok bool
	bedrock, ok = blocks.ConvertFromJava(input)
	if !ok {
		panic("找不到 Java 方块的转换")
	}
	fmt.Println(bedrock)
	// 2. 结果已整理；哈希反查是目标集合身份查询，不是恢复原始 Java 输入。
	var restored ir.IRBlock
	restored, ok = blocks.BedrockHashToBlockStrict(bedrock.Hash())
	fmt.Println(restored == bedrock, ok)
	// Output:
	// minecraft:jungle_fence_gate ["in_wall_bit"=false,"minecraft:cardinal_direction"="north","open_bit"=true]
	// true true
}

func ExampleConvertFromSchematic() {
	// 旧 Schematic 的 252 表示混凝土粉末，meta 12 选择棕色。
	const concretePowderID uint16 = 252
	const brownMeta uint16 = 12
	var block ir.IRBlock
	var ok bool
	block, ok = blocks.ConvertFromSchematic(concretePowderID, brownMeta)
	fmt.Println(block, ok)
	// Output: minecraft:brown_concrete_powder [] true
}
