package v1_test

import (
	"fmt"
	"neomega-blocks/ir"
	blocks "neomega-blocks/presets/block-converts/v1"
)

func ExampleConvertFromJava() {
	input := ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`)
	bedrock, ok := blocks.ConvertFromJava(input)
	if !ok {
		panic("unsupported Java block")
	}
	fmt.Println(bedrock)
	restored, ok := blocks.BedrockHashToBlockStrict(bedrock.Hash())
	fmt.Println(restored == bedrock, ok)
	// Output:
	// minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=true]
	// true true
}

func ExampleConvertFromSchematic() {
	block, ok := blocks.ConvertFromSchematic(252, 12)
	fmt.Println(block, ok)
	// Output: minecraft:brown_concrete_powder [] true
}
