package ir

import (
	decoder_shim "neomega-nbt/decoder-shim"
	decoder_snbt "neomega-nbt/decoder-snbt"
	encoder_snbt "neomega-nbt/encoder-snbt"
	"testing"
)

func TestBedrockPropsDecode(t *testing.T) {
	assertSnbt := func(propStr string, snbtStr string) {
		enc, getter := encoder_snbt.NewEncoderToBytes(nil)
		err := IRStates(propStr).DecodeToHL(enc)
		if err != nil {
			t.Fatalf("on decoding %v, get error: %v", propStr, err)
		}
		outStr := string(getter())
		if outStr != snbtStr {
			t.Fatalf("on decoding: %v, expect: %v, get: %v", propStr, snbtStr, outStr)
		}
	}
	assertSnbt(`[]`, `{}`)
	assertSnbt(`[""=""]`, `{"": ""}`)
	assertSnbt(`[  ""  =  ""  ]`, `{"": ""}`)
	assertSnbt(`[""="",]`, `{"": ""}`)
	assertSnbt(`[""="",  "1" =""]`, `{"": "", "1": ""}`)
	assertSnbt(`["strKey"="someStr",]`, `{"strKey": "someStr"}`)
	assertSnbt(`["trueKey"=  true ,]`, `{"trueKey": 1b}`)
	assertSnbt(`["trueKey"=  true ,"falseKey"=false]`, `{"trueKey": 1b, "falseKey": 0b}`)
	assertSnbt(`["intKey"=  1 ,]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"=  321 ,]`, `{"intKey": 321}`)
	assertSnbt(`["intKey"=  1,]`, `{"intKey": 1}`)
	assertSnbt(`["intKey"= 1]`, `{"intKey": 1}`)
	assertSnbt(`["strKey"="someStr","trueKey"=  true ,"falseKey"=false,"intKey1"=  1 ,"intKey321"=  321]`, `{"strKey": "someStr", "trueKey": 1b, "falseKey": 0b, "intKey1": 1, "intKey321": 321}`)
}

func TestDecoder(t *testing.T) {
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
		rec, getter := encoder_snbt.NewEncoderToBytes(nil)
		enc.Block().DecodeToHL(rec)
		out := string(getter())
		if out != expect {
			t.Fatalf("on testing %v, expect: %v, get: %v", snbt, expect, out)
		}
	}

	assertStr(
		`{"name":"minecraft:cyan_terracotta","states":{}}`,
		`{"name": "minecraft:cyan_terracotta", "states": {}}`,
	)
	assertStr(
		`{"name":"minecraft:polished_basalt","states":{"pillar_axis":"y"}}`,
		`{"name": "minecraft:polished_basalt", "states": {"pillar_axis": "y"}}`,
	)
	assertStr(
		`{"name":"minecraft:warped_door","states":{"door_hinge_bit":1b,"pillar_axis":"y","direction":1}}`,
		`{"name": "minecraft:warped_door", "states": {"direction": 1, "door_hinge_bit": 1b, "pillar_axis": "y"}}`,
	)
}

func TestHash(t *testing.T) {
	block, err := NewIRBlockFromGo("minecraft:cyan_terracotta", nil)
	if err != nil {
		t.Fatal(err)
	}
	if block.Hash() != 973836165 {
		t.FailNow()
	}

	block, err = NewIRBlockFromGo("minecraft:polished_basalt", map[string]any{
		"pillar_axis": "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	if block.Hash() != 2090385247 {
		t.FailNow()
	}

	block, err = NewIRBlockFromGo("minecraft:warped_door", map[string]any{
		"direction":       int32(0),
		"door_hinge_bit":  int8(0),
		"open_bit":        int8(0),
		"upper_block_bit": int8(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if block.Hash() != 4228134727 {
		t.FailNow()
	}
}
