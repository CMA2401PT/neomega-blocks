package v1

import (
	"neomega-blocks/ir"
	"testing"
)

func TestNEMCBedrockHashGet(t *testing.T) {
	blk, _ := NEMCBedrockHashToBlockStrict(3690217760)
	if blk != "minecraft:air []" {
		t.FailNow()
	}
	blk, _ = NEMCBedrockHashToBlockStrict(4107799631)
	if blk != `minecraft:bed ["direction"=3,"head_piece_bit"=false,"occupied_bit"=true]` {
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

func TestNEMCBedrockCanon(t *testing.T) {
	blk := ir.IRBlock(`minecraft:bed ["occupied_bit"=true,"direction"=3,"extraKey"="???"]`)
	out, _ := NEMCBedrockCanon(blk)
	if out != `minecraft:bed ["direction"=3,"head_piece_bit"=false,"occupied_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:notexist ["occupied_bit"=true,"direction"=3,"extraKey"="???"]`)
	out, _ = NEMCBedrockCanon(blk)
	if out != `minecraft:notexist ["direction"=3,"extraKey"="???","occupied_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:bed ["occupied_bit"=true,"direction"=3,"extraKey"="???"`)
	out, _ = NEMCBedrockCanon(blk)
	if out != `minecraft:bed ["direction"=0,"head_piece_bit"=false,"occupied_bit"=false]` {
		t.Fatalf("output= %v", out)
	}
}

func TestConvertBedrockToNEMCBedrock(t *testing.T) {
	blk := ir.IRBlock(`minecraft:tripWire ["powered_bit"=true]`)
	out, _ := ConvertBedrockToNEMCBedrock(blk)
	if out != `minecraft:trip_wire ["attached_bit"=false,"disarmed_bit"=false,"powered_bit"=true,"suspended_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:dropper ["facing_direction"=0]`)
	out, _ = ConvertBedrockToNEMCBedrock(blk)
	if out != `minecraft:dropper ["facing_direction"=0,"triggered_bit"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:notexist ["occupied_bit"=true,"direction"=3,"extraKey"="???"`)
	out, _ = ConvertBedrockToNEMCBedrock(blk)
	if out != `minecraft:unknown []` {
		t.Fatalf("output= %v", out)
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

func TestJavaConvert(t *testing.T) {
	blk := ir.IRBlock(`minecraft:jungle_fence_gate ["powered"="false","open"="false"]`)
	out, _ := ConvertJavaToNEMCBedrock(blk)
	if out != `minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=false]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:red_sandstone_wall ["south"="none","north"="none","east"="tall"]`)
	out, _ = ConvertJavaToNEMCBedrock(blk)
	if out != `minecraft:red_sandstone_wall ["wall_connection_type_east"="tall","wall_connection_type_north"="none","wall_connection_type_south"="none","wall_connection_type_west"="none","wall_post_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

	blk = ir.IRBlock(`minecraft:jungle_fence_gate ["open"="true"]`)
	out, _ = ConvertJavaToNEMCBedrock(blk)
	if out != `minecraft:jungle_fence_gate ["direction"=2,"in_wall_bit"=false,"open_bit"=true]` {
		t.Fatalf("output= %v", out)
	}

}
