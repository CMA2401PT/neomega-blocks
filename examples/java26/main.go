package main

import (
	"fmt"
	"log"
	"neomega-blocks/ir"
	java "neomega-blocks/presets/java/v26_3"
)

func main() {
	c, err := java.Open()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Java %s: %d blocks, %d full states\n", c.Metadata().JavaVersion, c.BlockCount(), c.StateCount())
	for _, name := range []string{"copper_chest", "sulfur", "cinnabar", "oak_slab"} {
		block, err := c.Normalize(ir.IRBlock(name + " []"))
		if err != nil {
			log.Fatal(err)
		}
		mapping, err := c.ToBedrock(block)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s -> %s (waterlogged=%t, reverse candidates=%d)\n", block, mapping.Block, mapping.Waterlogged, len(c.JavaCandidates(mapping.Block)))
	}
}
