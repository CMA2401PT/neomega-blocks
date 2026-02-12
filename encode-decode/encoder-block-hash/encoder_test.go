package encoder_block_hash

import (
	nbt "neomega-nbt"
	decoder_go_object "neomega-nbt/decoder-go-object"
	decoder_shim "neomega-nbt/decoder-shim"
	"testing"
)

func TestHash(t *testing.T) {
	enc := NewEncoder()
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("")
	enc.WriteType(nbt.TagString)
	enc.WriteTag("name")
	enc.WriteTag("minecraft:cyan_terracotta")
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("states")

	enc.WriteEnd()
	enc.WriteEnd()
	if enc.Hash() != 973836165 {
		t.FailNow()
	}

	enc.Reset()
	decoder_go_object.HighLevelDecodeTo(
		map[string]any{
			"name":   "minecraft:cyan_terracotta",
			"states": map[string]any{},
		},
		decoder_shim.NewDecoderShim(enc),
		decoder_go_object.DecodeOption{},
	)
	if enc.Hash() != 973836165 {
		t.FailNow()
	}

	enc.Reset()
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("")
	enc.WriteType(nbt.TagString)
	enc.WriteTag("name")
	enc.WriteTag("minecraft:polished_basalt")
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("states")
	enc.WriteType(nbt.TagString)
	enc.WriteTag("pillar_axis")
	enc.WriteString("y")
	enc.WriteEnd()
	enc.WriteEnd()
	if enc.Hash() != 2090385247 {
		t.FailNow()
	}

	enc.Reset()
	decoder_go_object.HighLevelDecodeTo(
		map[string]any{
			"name": "minecraft:polished_basalt",
			"states": map[string]any{
				"pillar_axis": "y",
			},
		},
		decoder_shim.NewDecoderShim(enc),
		decoder_go_object.DecodeOption{
			SortMapKey: true,
		},
	)
	if enc.Hash() != 2090385247 {
		t.FailNow()
	}

	enc.Reset()
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("")
	enc.WriteType(nbt.TagString)
	enc.WriteTag("name")
	enc.WriteTag("minecraft:warped_door")
	enc.WriteType(nbt.TagCompound)
	enc.WriteTag("states")
	enc.WriteType(nbt.TagInt)
	enc.WriteTag("direction")
	enc.WriteInt32(0)
	enc.WriteType(nbt.TagByte)
	enc.WriteTag("door_hinge_bit")
	enc.WriteInt8(0)
	enc.WriteType(nbt.TagByte)
	enc.WriteTag("open_bit")
	enc.WriteInt8(0)
	enc.WriteType(nbt.TagByte)
	enc.WriteTag("upper_block_bit")
	enc.WriteInt8(0)
	enc.WriteEnd()
	enc.WriteEnd()
	if enc.Hash() != 4228134727 {
		t.FailNow()
	}

	enc.Reset()
	decoder_go_object.HighLevelDecodeTo(
		map[string]any{
			"name": "minecraft:warped_door",
			"states": map[string]any{
				"direction":       int32(0),
				"door_hinge_bit":  int8(0),
				"open_bit":        int8(0),
				"upper_block_bit": int8(0),
			},
		},
		decoder_shim.NewDecoderShim(enc),
		decoder_go_object.DecodeOption{
			SortMapKey: true,
		},
	)
	if enc.Hash() != 4228134727 {
		t.FailNow()
	}

}
