package ir

import (
	decoder_shim "neomega-nbt/decoder-shim"
	decoder_snbt "neomega-nbt/decoder-snbt"
	"testing"
)

func TestBedrockPropsEncode(t *testing.T) {
	assertSnbt := func(propStr string, snbtStr string) {
		enc := NewIRStateEncoder()
		_, err := decoder_snbt.DecodeSnbtTo([]uint8(snbtStr), decoder_shim.NewDecoderShim(enc), decoder_snbt.DecodeOption{})
		if err != nil {
			t.Fatalf("on decoding %v, get error: %v", snbtStr, err)
		}
		outStr := string(enc.States())
		if outStr != propStr {
			t.Fatalf("on decoding: %v, expect: %v, get: %v", snbtStr, propStr, outStr)
		}
	}
	assertSnbt(`[]`, `{}`)
	assertSnbt(`["key1"=""]`, `{"key1": ""}`)
	assertSnbt(`[""="","1"=""]`, `{"": "", "1": ""}`)
	assertSnbt(`["trueKey"=true]`, `{"trueKey": 1b}`)
	assertSnbt(`["falseKey"=false,"trueKey"=true]`, `{"trueKey": 1b, "falseKey": 0b}`)
	assertSnbt(`["intKey"=0]`, `{"intKey": 0}`)
	assertSnbt(`["intKey"=1]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"=321]`, `{"intKey": 321}`)
	assertSnbt(`["falseKey"=false,"intKey1"=1,"intKey321"=321,"strKey"="someStr","trueKey"=true]`, `{"strKey": "someStr", "trueKey": 1b, "falseKey": 0b, "intKey1": 1, "intKey321": 321}`)
}

func TestEncoder(t *testing.T) {
	assertStr := func(snbt string, expect string) {
		enc := &BlockEncoder{}
		_, err := decoder_snbt.DecodeSnbtTo(
			[]uint8(snbt),
			decoder_shim.NewDecoderShim(enc),
			decoder_snbt.DecodeOption{},
		)
		if err != nil {
			t.Fatalf("on testing %v, get error: %v", snbt, err)
		}
		if enc.Block().String() != expect {
			t.Fatalf("on testing %v, expect: %v, get: %v", snbt, expect, enc.Block().String())
		}
	}

	assertStr(`{"name":"minecraft:cyan_terracotta","states":{}}`, `minecraft:cyan_terracotta []`)
	assertStr(
		`{"name":"minecraft:polished_basalt","states":{"pillar_axis":"y"}}`,
		`minecraft:polished_basalt ["pillar_axis"="y"]`,
	)
	assertStr(
		`{"name":"minecraft:warped_door","states":{"door_hinge_bit":1b,"pillar_axis":"y","direction":1}}`,
		`minecraft:warped_door ["direction"=1,"door_hinge_bit"=true,"pillar_axis"="y"]`,
	)
}

func TestEncoderGo(t *testing.T) {
	assertStr := func(name string, props map[string]any, expect string) {
		block, err := NewIRBlockFromGo(name, props)
		if err != nil {
			t.Fatalf("on testing %v, get error: %v", props, err)
		}
		if string(block) != expect {
			t.Fatalf("on testing %v, expect: %v, get: %v", props, expect, string(block))
		}
	}
	assertStr("minecraft:polished_basalt", map[string]any{}, `minecraft:polished_basalt []`)
	assertStr("minecraft:polished_basalt", nil, `minecraft:polished_basalt []`)
	assertStr("minecraft:polished_basalt", map[string]any{
		"pillar_axis": "y",
	}, `minecraft:polished_basalt ["pillar_axis"="y"]`)
	assertStr("minecraft:warped_door", map[string]any{
		"open_bit":        int8(0),
		"door_hinge_bit":  int8(1),
		"upper_block_bit": int8(0),
		"direction":       int32(3),
	}, `minecraft:warped_door ["direction"=3,"door_hinge_bit"=true,"open_bit"=false,"upper_block_bit"=false]`)
}
