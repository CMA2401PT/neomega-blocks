package ir_go

// import (
// 	decoder_shim "neomega-nbt/decoder-shim"
// 	decoder_snbt "neomega-nbt/decoder-snbt"
// 	encoder_snbt "neomega-nbt/encoder-snbt"
// 	"testing"
// )

// func TestName(t *testing.T) {
// 	b := IRGoBlock{
// 		nameWithNameSpace: "minecraft:block",
// 	}
// 	if b.NameSpace() != "minecraft" {
// 		t.FailNow()
// 	}
// 	if b.ShortName() != "block" {
// 		t.FailNow()
// 	}
// 	b = IRGoBlock{
// 		nameWithNameSpace: "block",
// 	}
// 	if b.NameSpace() != "" {
// 		t.FailNow()
// 	}
// 	if b.ShortName() != "block" {
// 		t.FailNow()
// 	}
// }

// func TestEncoder(t *testing.T) {
// 	assertStr := func(snbt string, expect string) {
// 		enc := &Encoder{}
// 		_, err := decoder_snbt.DecodeSnbtTo(
// 			[]uint8(snbt),
// 			decoder_shim.NewDecoderShim(enc),
// 			decoder_snbt.DecodeOption{},
// 		)
// 		if err != nil {
// 			t.Fatalf("on testing %v, get error: %v", snbt, err)
// 		}
// 		if enc.block.String() != expect {
// 			t.Fatalf("on testing %v, expect: %v, get: %v", snbt, expect, enc.block.String())
// 		}
// 	}

// 	assertStr(`{"name":"minecraft:cyan_terracotta","states":{}}`, `minecraft:cyan_terracotta []`)
// 	assertStr(
// 		`{"name":"minecraft:polished_basalt","states":{"pillar_axis":"y"}}`,
// 		`minecraft:polished_basalt ["pillar_axis"="y"]`,
// 	)
// 	assertStr(
// 		`{"name":"minecraft:warped_door","states":{"door_hinge_bit":1b,"pillar_axis":"y","direction":1}}`,
// 		`minecraft:warped_door ["direction"=1,"door_hinge_bit"=true,"pillar_axis"="y"]`,
// 	)
// }

// func TestEncoderGo(t *testing.T) {
// 	assertStr := func(name string, props map[string]any, expect string) {
// 		block, err := NewIRBlockFromGo(name, props)
// 		if err != nil {
// 			t.Fatalf("on testing %v, get error: %v", props, err)
// 		}
// 		if block.String() != expect {
// 			t.Fatalf("on testing %v, expect: %v, get: %v", props, expect, block.String())
// 		}
// 	}
// 	assertStr("minecraft:polished_basalt", map[string]any{}, `minecraft:polished_basalt []`)
// 	assertStr("minecraft:polished_basalt", nil, `minecraft:polished_basalt []`)
// 	assertStr("minecraft:polished_basalt", map[string]any{
// 		"pillar_axis": "y",
// 	}, `minecraft:polished_basalt ["pillar_axis"="y"]`)
// 	assertStr("minecraft:warped_door", map[string]any{
// 		"open_bit":        int8(0),
// 		"door_hinge_bit":  int8(1),
// 		"upper_block_bit": int8(0),
// 		"direction":       int32(3),
// 	}, `minecraft:warped_door ["direction"=3,"door_hinge_bit"=true,"open_bit"=false,"upper_block_bit"=false]`)
// }

// func TestDecoder(t *testing.T) {
// 	assertStr := func(snbt string, expect string) {
// 		enc := &Encoder{}
// 		_, err := decoder_snbt.DecodeSnbtTo(
// 			[]uint8(snbt),
// 			decoder_shim.NewDecoderShim(enc),
// 			decoder_snbt.DecodeOption{},
// 		)
// 		if err != nil {
// 			t.Fatalf("on testing %v, get error: %v", snbt, err)
// 		}
// 		rec, getter := encoder_snbt.NewEncoderToBytes(nil)
// 		enc.block.DecodeToHighLevel(rec)
// 		out := string(getter())
// 		if out != expect {
// 			t.Fatalf("on testing %v, expect: %v, get: %v", snbt, expect, out)
// 		}
// 	}

// 	assertStr(
// 		`{"name":"minecraft:cyan_terracotta","states":{}}`,
// 		`{"name": "minecraft:cyan_terracotta", "states": {}}`,
// 	)
// 	assertStr(
// 		`{"name":"minecraft:polished_basalt","states":{"pillar_axis":"y"}}`,
// 		`{"name": "minecraft:polished_basalt", "states": {"pillar_axis": "y"}}`,
// 	)
// 	assertStr(
// 		`{"name":"minecraft:warped_door","states":{"door_hinge_bit":1b,"pillar_axis":"y","direction":1}}`,
// 		`{"name": "minecraft:warped_door", "states": {"direction": 1, "door_hinge_bit": 1b, "pillar_axis": "y"}}`,
// 	)
// }
