// 在 neomega-blocks 目录运行：go run ./examples/convert
package main

import (
	"fmt"

	"neomega-blocks/ir"
	blocks "neomega-blocks/presets/block-converts/v1"
)

func main() {
	// 1. IRBlock 用文本描述方块名和属性；Java 的 open 是字符串 "true"。
	// 先 Tidy 统一命名空间和属性顺序；省略的游戏属性由转换入口补默认值。
	var javaInput ir.IRBlock = ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`).Tidy()
	var bedrock ir.IRBlock
	var found bool
	bedrock, found = blocks.ConvertFromJava(javaInput)
	// 转换是否成功看 bool；失败时返回的 unknown 方块不能当作有效结果。
	if !found {
		panic("找不到 Java 方块的转换")
	}
	fmt.Println(bedrock)

	// 2. 成功返回的方块已经整理，可以直接取目标 palette 的网络哈希并反查。
	// 这是目标集合内的身份查询，不是把 Bedrock 还原成原始 Java 输入。
	var hash uint32 = bedrock.Hash()
	var restored ir.IRBlock
	restored, found = blocks.BedrockHashToBlockStrict(hash)
	if !found {
		panic("找不到目标方块的网络哈希")
	}
	fmt.Printf("哈希反查一致：%t\n", restored == bedrock)

	// 3. 旧 Schematic 用数字 ID 和 meta 描述方块；252 的 meta 12 是棕色混凝土粉末。
	const concretePowderID uint16 = 252
	const brownMeta uint16 = 12
	var schematicBlock ir.IRBlock
	schematicBlock, found = blocks.ConvertFromSchematic(concretePowderID, brownMeta)
	if !found {
		panic("找不到旧 Schematic 方块的转换")
	}
	fmt.Println(schematicBlock)
}
