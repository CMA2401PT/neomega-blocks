package v1

import (
	"neomega-blocks/ir"
	"testing"
)

func TestNEMCBedrockHashGet(t *testing.T) {
	blk, _ := BedrockHashToBlockStrict(3690217760)
	if blk != "minecraft:air []" {
		t.FailNow()
	}
	var ok bool
	_, ok = BedrockHashToBlockStrict(23768)
	if ok {
		t.FailNow()
	}
	blk, _ = BedrockHashToBlockStrict(4107799631)
	if blk != `minecraft:bed ["direction"=3,"head_piece_bit"=false,"occupied_bit"=true]` {
		t.FailNow()
	}

	blk, ok = BedrockHashToBlockStrict(ir.IRBlock(`minecraft:nether_wart ["age"=3]`).Hash())
	if blk != `minecraft:nether_wart ["age"=3]` {
		t.FailNow()
	}
}

func TestNEMCBedrockNameToDefaultProps(t *testing.T) {
	props, ok := nemcBedrockNameToDefaultProps("minecraft:bed")
	if !ok {
		t.Fatal("expected default props for minecraft:bed")
	}
	if props != ir.IRStates(`["direction"=0,"head_piece_bit"=false,"occupied_bit"=false]`) {
		t.Fatalf("unexpected default props for bed: %q", props)
	}
}

func TestBedrockDataValueToHash(t *testing.T) {
	hash, ok := BedrockDataValueToHash("air", 0)
	if !ok {
		t.Fatal("air data=0 should be found")
	}
	if hash != 3690217760 {
		t.Fatalf("air data=0 hash=%d, want 3690217760", hash)
	}

	hash, ok = BedrockDataValueToHash("minecraft:air", 9999)
	if ok || hash != 0 {
		t.Fatalf("air data out of range should be (0,false), got (%d,%v)", hash, ok)
	}

	hash, ok = BedrockDataValueToHash("polished_blackstone_button", 6)
	if ok || hash != 0 {
		t.Fatalf("hole data should be (0,false), got (%d,%v)", hash, ok)
	}

	hash, ok = BedrockDataValueToHash("birch_standing_sign", 2)
	if hash != 1099896350 {
		t.Fatalf("should be 1099896350, got (%d,%v)", hash, ok)
	}

	blk, ok := BedrockDataValueToBlock("birch_standing_sign", 2)
	if blk.String() != `minecraft:birch_standing_sign ["ground_sign_direction"=2]` {
		t.Fatalf("got (%v,%v)", blk, ok)
	}

	name, data, ok := BedrockHashToBlockData(ir.IRBlock(`minecraft:birch_standing_sign ["ground_sign_direction"=2]`).Hash())
	if name != `minecraft:birch_standing_sign` || data != 2 || !ok {
		t.FailNow()
	}
}

func TestNEMCBedrockCanon(t *testing.T) {
	blk := ir.IRBlock(`minecraft:bed ["occupied_bit"=true,"direction"=3,"extraKey"="???"]`)
	out, _ := BedrockCanon(blk)
	if out != `minecraft:bed ["direction"=3,"head_piece_bit"=false,"occupied_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:notexist ["occupied_bit"=true,"direction"=3,"extraKey"="???"]`)
	out, _ = BedrockCanon(blk)
	if out != `minecraft:notexist ["direction"=3,"extraKey"="???","occupied_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:bed ["occupied_bit"=true,"direction"=3,"extraKey"="???"`)
	out, _ = BedrockCanon(blk)
	if out != `minecraft:bed ["direction"=0,"head_piece_bit"=false,"occupied_bit"=false]` {
		t.Fatalf("output= %v", out)
	}
}

func TestConvertBedrockToNEMCBedrock(t *testing.T) {
	blk, _ := ConvertFromBedrock(ir.IRBlock(`minecraft:nether_wart ["age"="3"]`))
	if blk != `minecraft:nether_wart ["age"=3]` {
		t.FailNow()
	}
	blk = ir.IRBlock(`minecraft:tripWire ["powered_bit"=true]`)
	out, _ := ConvertFromBedrock(blk)
	if out != `minecraft:trip_wire ["attached_bit"=false,"disarmed_bit"=false,"powered_bit"=true,"suspended_bit"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:dropper ["facing_direction"=0]`)
	out, _ = ConvertFromBedrock(blk)
	if out != `minecraft:dropper ["facing_direction"=0,"triggered_bit"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:notexist ["occupied_bit"=true,"direction"=3,"extraKey"="???"`)
	out, _ = ConvertFromBedrock(blk)
	if out != `minecraft:unknown []` {
		t.Fatalf("output= %v", out)
	}

	// 旧属性需要查表保留语义，缺参输入仍应由 Canon 补齐默认值。
	for _, tc := range []struct {
		name  string
		input ir.IRBlock
		want  ir.IRBlock
	}{
		{
			name:  "stone_granite",
			input: `minecraft:stone ["stone_type"="granite"]`,
			want:  `minecraft:granite []`,
		},
		{
			name:  "red_sand",
			input: `minecraft:sand ["sand_type"="red"]`,
			want:  `minecraft:red_sand []`,
		},
		{
			name:  "furnace_east",
			input: `minecraft:furnace ["facing_direction"=5]`,
			want:  `minecraft:furnace ["custom_appearance"=false,"minecraft:cardinal_direction"="east"]`,
		},
		{
			name:  "bamboo_slab_top",
			input: `minecraft:bamboo_slab ["top_slot_bit"=true]`,
			want:  `minecraft:bamboo_slab ["minecraft:vertical_half"="top"]`,
		},
		{
			name:  "red_nether_brick_wall",
			input: `minecraft:cobblestone_wall ["wall_block_type"="red_nether_brick"]`,
			want:  `minecraft:red_nether_brick_wall ["wall_connection_type_east"="none","wall_connection_type_north"="none","wall_connection_type_south"="none","wall_connection_type_west"="none","wall_post_bit"=false]`,
		},
		{
			name:  "acacia_door_defaults",
			input: `minecraft:acacia_door ["upper_block_bit"=true]`,
			want:  `minecraft:acacia_door ["direction"=0,"door_hinge_bit"=false,"open_bit"=false,"upper_block_bit"=true]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ConvertFromBedrock(tc.input)
			if !ok || got != tc.want {
				t.Fatalf("ConvertFromBedrock(%v) = (%v, %v), want (%v, true)", tc.input, got, ok, tc.want)
			}
		})
	}
}

func TestConvertFromBedrockTargetStates(t *testing.T) {
	// 目标状态必须优先直接命中，不能被兼容旧输入的映射覆盖。
	for _, block := range BedrockBlocks() {
		got, ok := ConvertFromBedrock(block)
		if !ok || got != block {
			t.Fatalf("ConvertFromBedrock(%v) = (%v, %v), want (%v, true)", block, got, ok, block)
		}
	}
}

func TestJavaCanon(t *testing.T) {
	blk := ir.IRBlock(`minecraft:bed ["facing"="north","occupied"="false","extraKey"="???"]`)
	out, _ := JavaCanon(blk)
	if out != `minecraft:bed ["color"="red","facing"="north","occupied"="false","part"="foot"]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:notexist ["occupied_bit"=true,"direction"=3,"extraKey"="???"]`)
	out, _ = JavaCanon(blk)
	if out != `minecraft:notexist ["direction"=3,"extraKey"="???","occupied_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:bed ["occupied_bit"=true,"direction"=3,"extraKey"="???"`)
	out, _ = JavaCanon(blk)
	if out != `minecraft:bed ["color"="red","facing"="north","occupied"="false","part"="foot"]` {
		t.Fatalf("output= %v", out)
	}
}

func TestConvertFromJava(t *testing.T) {
	blk := ir.IRBlock(`minecraft:jungle_fence_gate   [ "powered"="false",  "open"="false"]`)
	out, _ := ConvertFromJava(blk)
	if out != `minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:orange_glazed_terracotta ["facing"="east"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:orange_glazed_terracotta ["facing_direction"=5]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:cobblestone_wall    [ "up"="true" ,  "west"="tall","south"="short","east"="none"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:cobblestone_wall ["wall_connection_type_east"="none","wall_connection_type_north"="none","wall_connection_type_south"="short","wall_connection_type_west"="tall","wall_post_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:cobblestone_wall    [ "up"="true" ,  "west"="tall","south"="low","east"="none"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:cobblestone_wall ["wall_connection_type_east"="none","wall_connection_type_north"="none","wall_connection_type_south"="short","wall_connection_type_west"="tall","wall_post_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:birch_wall_hanging_sign ["facing"="west"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:birch_hanging_sign ["attached_bit"=false,"facing_direction"=4,"ground_sign_direction"=0,"hanging"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:mangrove_hanging_sign ["rotation"="10"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:mangrove_hanging_sign ["attached_bit"=false,"facing_direction"=0,"ground_sign_direction"=10,"hanging"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:red_sandstone_slab ["type"="double"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:red_sandstone_double_slab ["minecraft:vertical_half"="bottom"]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:red_sandstone_wall ["south"="none","north"="none","east"="tall"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:red_sandstone_wall ["wall_connection_type_east"="tall","wall_connection_type_north"="none","wall_connection_type_south"="none","wall_connection_type_west"="none","wall_post_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=true]` {
		t.Fatalf("output= %v", out)
	}
	blk = ir.IRBlock(`minecraft:grass_block []`)
	out, _ = ConvertFromJava(blk)
	if out != `minecraft:grass_block []` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:cauldron ["level"="3"]`)
	out, ok := ConvertFromJava(blk)
	if !ok || out != `minecraft:cauldron ["cauldron_liquid"="water","fill_level"=6]` {
		t.Fatalf("output= %v, ok= %v", out, ok)
	}
}

func TestConvertToJava(t *testing.T) {
	blk := ir.IRBlock(`minecraft:golden_rail [ "rail_direction"=9]`)
	out, _ := ConvertToJava(blk)
	if out != `minecraft:powered_rail ["powered"="false","shape"="north_south"]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:orange_glazed_terracotta ["facing_direction"=5]`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:orange_glazed_terracotta ["facing"="east"]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:frog_spawn []`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:frogspawn []` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:grass_path []`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:dirt_path []` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:cobblestone_wall ["wall_connection_type_south"="short","wall_connection_type_west"="tall"]`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:cobblestone_wall ["east"="none","north"="none","south"="low","up"="false","west"="tall"]` {
		t.Fatalf("output= %v", out)
	}
	blk = ir.IRBlock(`minecraft:orange_glazed_terracotta ["facing_direction"=5]`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:orange_glazed_terracotta ["facing"="east"]` {
		t.Fatalf("output= %v", out)
	}
	blk = ir.IRBlock(`minecraft:birch_hanging_sign ["facing_direction"=4,"ground_sign_direction"=0,"hanging"=false]`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:birch_wall_hanging_sign ["facing"="west"]` {
		t.Fatalf("output= %v", out)
	}
	blk = ir.IRBlock(`minecraft:birch_hanging_sign ["facing_direction"=0,"ground_sign_direction"=10,"hanging"=true]`)
	out, _ = ConvertToJava(blk)
	if out != `minecraft:birch_hanging_sign ["attached"="false","rotation"="10"]` {
		t.Fatalf("output= %v", out)
	}
}

func TestConvertFromSchematic(t *testing.T) {
	out, _ := ConvertFromSchematic(252, 12)
	if out != `minecraft:brown_concrete_powder []` {
		t.Fatalf("output= %v", out)
	}
	out, _ = ConvertFromSchematic(154, 4)
	if out != `minecraft:hopper ["facing_direction"=4,"toggle_bit"=true]` {
		t.Fatalf("output= %v", out)
	}
	out, _ = ConvertFromSchematic(180, 2)
	if out != `minecraft:red_sandstone_stairs ["upside_down_bit"=false,"weirdo_direction"=2]` {
		t.Fatalf("output= %v", out)
	}
	out, _ = ConvertFromSchematic(2, 0)
	if out != `minecraft:grass_block []` {
		t.Fatalf("output= %v", out)
	}
}
