//go:build resourcegen

package blockconverts

import (
	"bufio"
	"bytes"
	"fmt"
	"neomega-blocks/ir"
	specific_converts "neomega-blocks/presets/block-converts/specific-converts"
	presets "neomega-blocks/presets/block-converts/v1"
	blocks_name "neomega-blocks/storage/blocks-name"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func LoadBlockList(file string) []ir.IRBlock {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(bytes.NewReader(data))
	blocks := []ir.IRBlock{}
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineStr := strings.TrimSpace(string(line))
		blocks = append(blocks, ir.IRBlock(lineStr).Tidy())
	}
	fmt.Println(len(blocks))
	return blocks
}

func bedrockWallDirectConvert(inBedrock ir.IRBlock) (outJava ir.IRBlock, ok bool) {
	if !strings.HasSuffix(inBedrock.NameWithNameSpace(), "_wall") {
		return "", false
	}
	inBedrock, ok = presets.BedrockCanon(inBedrock)
	if !ok {
		return "", false
	}
	outStates := inBedrock.States().Update(specific_converts.BedrockToJavaWallRelplacer)
	return ir.IRBlock(inBedrock.NameWithNameSpace() + " " + string(outStates)), true
}
func bedrockHangingSignDirectConvert(inBedrock ir.IRBlock) (outJava ir.IRBlock, ok bool) {
	if outJava, ok := specific_converts.BedrockToJavaHangingSign(inBedrock); ok {
		return outJava, ok
	}
	return inBedrock, false
}

func TestStep3GenJavaFanout(t *testing.T) {
	allowsList := LoadBlockList("tmp/fan_out_java_allowed.list")
	allows := map[ir.IRBlock]struct{}{}
	for _, b := range allowsList {
		allows[b] = struct{}{}
	}
	isAllow := func(b ir.IRBlock) bool {
		_, ok := allows[b]
		return ok
	}
	pairs := LoadMappingPairs("tmp/fan_out_java.mapping")

	fullPairs := map[ir.IRBlock]ir.IRBlock{}
	tidyPairs := []MappingPair{}

	for _, p := range pairs {
		outJava, ok := bedrockWallDirectConvert(p.In)
		if ok {
			if !isAllow(outJava) {
				t.Errorf("direct convert %v not in java block space", outJava)
			}
			continue
		}
		outJava, ok = bedrockHangingSignDirectConvert(p.In)
		if ok {
			if !isAllow(outJava) {
				t.Errorf("direct convert %v not in java block space", outJava)
			}
			continue
		}
		if !isAllow(p.Out) {
			t.Errorf("output %v not in java block space", p.Out.String())
		}
		cannonIn, ok := presets.BedrockCanon(p.In)
		if !ok {
			panic("non cannon bedrock block exist")
		}
		if p.In == cannonIn {
			fullPairs[p.In] = p.Out
			tidyPairs = append(tidyPairs, p)
		}
	}

	directConvert := 0
	ignorePairs := 0
	nonCannonPairs := 0
	for _, p := range pairs {
		_, ok := bedrockWallDirectConvert(p.In)
		if ok {
			directConvert += 1
			continue
		}
		_, ok = bedrockHangingSignDirectConvert(p.In)
		if ok {
			directConvert += 1
			continue
		}

		cannonIn, ok := presets.BedrockCanon(p.In)
		if !ok {
			t.FailNow()
		}
		_, found := fullPairs[cannonIn]
		if found {
			ignorePairs += 1
		} else {
			nonCannonPairs += 1
			tidyPairs = append(tidyPairs, p)
		}
	}
	fmt.Println(len(fullPairs))
	fmt.Printf("skip %v pairs because cannon, skip %v pairs because direct, %v (%v non cannon) remaining\n", ignorePairs, directConvert, len(tidyPairs), nonCannonPairs)
	outStr := ""
	for _, p := range tidyPairs {
		outStr += fmt.Sprintf("%v#%v\n", p.In.String(), p.Out.String())
	}
	if err := os.WriteFile("tmp/fan_out_java_tidy.mapping", []byte(outStr), 0644); err != nil {
		t.Fatal(err)
	}

	storage := blocks_name.NewBlockNameStorage[uint16]()
	for _, p := range tidyPairs {
		storage.SetWithSpecificHash(p.Out, blocks_name.BlockHash(p.In.Hash()))
	}
	storage.SetWithSpecificHash("minecraft:air []", blocks_name.BlockHash(ir.IRBlock("minecraft:micro_block []").Hash()))
	dumped := storage.Dump()
	loaded := blocks_name.LoadBlockNameStorage[uint16](dumped)
	_, cur, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	palettePath := filepath.Join(filepath.Dir(cur), "block_palette.json")
	for _, b := range ReadBlocks(palettePath) {
		_, ok := bedrockWallDirectConvert(b.IRBlock)
		if ok {
			continue
		}
		_, ok = bedrockHangingSignDirectConvert(b.IRBlock)
		if ok {
			continue
		}
		_, ok = loaded.Get(blocks_name.BlockHash(b.IRBlock.Hash()))
		if !ok {
			t.Fatalf("%v cannot convert to java", b.IRBlock)
		}
	}
	fmt.Printf("%v MiB\n", float32(len(storage.Dump()))/(1<<20))
	if err := os.WriteFile("../../../presets/block-converts/v1/3_to_java.bin", storage.Dump(), 0644); err != nil {
		t.Fatal(err)
	}

}
