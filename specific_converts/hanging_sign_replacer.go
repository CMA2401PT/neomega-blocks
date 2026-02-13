package specific_converts

import (
	"fmt"
	"neomega-blocks/ir"
	"strconv"
	"strings"
)

func facingToInt(facing any) int32 {
	switch facing {
	case "east":
		return int32(5)
	case "west":
		return int32(4)
	case "south":
		return int32(3)
	case "north":
		return int32(2)
	case "up":
		return int32(1)
	case "down":
		return int32(0)
	default:
		return int32(0)
	}
}

func intToFacing(direction any) string {
	switch direction {
	case int32(5):
		return "east"
	case int32(4):
		return "west"
	case int32(3):
		return "south"
	case int32(2):
		return "north"
	case int32(1):
		return "up"
	case int32(0):
		return "down"
	default:
		return "north"
	}
}

func intToFacing4(direction any) string {
	switch direction {
	case int32(5):
		return "east"
	case int32(4):
		return "west"
	case int32(3):
		return "south"
	case int32(2):
		return "north"
	case int32(1):
		return "north"
	case int32(0):
		return "north"
	default:
		return "north"
	}
}

func BedrockToJavaHangingSign(bedrockBlock ir.IRBlock) (ir.IRBlock, bool) {
	if strings.HasSuffix(bedrockBlock.ShortName(), "_hanging_sign") {
		// 其关键在于，有一个 hanging bit，决定其究竟是哪一种 java block
		bedrockStates := bedrockBlock.States()
		signMaterial := strings.TrimSuffix(bedrockBlock.ShortName(), "_hanging_sign")
		if bedrockStates.CheckKey("hanging", true) {
			// ground_sign_direction -> "rotation" attached_bit->"true"/"false"
			rotation := bedrockStates.GetDefault("ground_sign_direction", int32(0))
			attachStr := "false"
			if bedrockStates.CheckKey("attached_bit", true) {
				attachStr = "true"
			}
			return ir.IRBlock(fmt.Sprintf(`minecraft:%v_hanging_sign ["attached"="%v","rotation"="%v"]`, signMaterial, attachStr, rotation)), true
		} else {
			// wall
			facingInt := bedrockStates.GetDefault("facing_direction", int32(2))
			return ir.IRBlock(fmt.Sprintf(`minecraft:%v_wall_hanging_sign ["facing"="%v"]`, signMaterial, intToFacing4(facingInt))), true
		}
	}
	return bedrockBlock, false
}

func JavaToBedrockHangingSign(javaBlock ir.IRBlock) (ir.IRBlock, bool) {
	if strings.HasSuffix(javaBlock.ShortName(), "_wall_hanging_sign") {
		states := javaBlock.States()
		signMaterial := strings.TrimSuffix(javaBlock.ShortName(), "_wall_hanging_sign")
		out, err := ir.NewIRBlockFromGo(
			fmt.Sprintf(`minecraft:%v_hanging_sign`, signMaterial),
			map[string]any{
				"hanging":               false,
				"attached_bit":          states.CheckKey("attached", "true"),
				"ground_sign_direction": int32(0),
				"facing_direction":      facingToInt(states.GetDefault("facing", "north")),
			},
		)
		if err == nil {
			return out, true
		}
	} else if strings.HasSuffix(javaBlock.ShortName(), "_hanging_sign") {
		states := javaBlock.States()
		signMaterial := strings.TrimSuffix(javaBlock.ShortName(), "_hanging_sign")
		rotation := states.GetDefault("rotation", "0")
		rotationS, ok := rotation.(string)
		if !ok {
			rotationS = "0"
		}
		ret, _ := strconv.Atoi(rotationS)
		out, err := ir.NewIRBlockFromGo(
			fmt.Sprintf(`minecraft:%v_hanging_sign`, signMaterial),
			map[string]any{
				"hanging":               true,
				"attached_bit":          states.CheckKey("attached", "true"),
				"ground_sign_direction": int32(ret),
				"facing_direction":      int32(0),
			},
		)
		if err == nil {
			return out, true
		}
	}
	return javaBlock, false
}

// 	// hbi, ok := bedrockStates.Get("hanging")
// 	// isHanging := false
// 	// if ok {
// 	// 	isHangingU8, _ := hbi.(uint8)
// 	// 	isHanging = isHangingU8 == 1
// 	// }
// 	// if isHanging {
// 	// 	blockName=
// 	// }
// }

// east:5 west:4 south: 3 north:2 up:1 ?down:0?
