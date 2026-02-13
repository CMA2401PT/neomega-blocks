package specific_converts

import "strings"

// 只考虑当前版本
func BedrockToJavaWallRelplacer(key string, value any) (string, any) {
	if key == "wall_post_bit" {
		vout := "false"
		vin, _ := value.(uint8)
		if vin == 1 {
			vout = "true"
		}
		return "up", vout
	}
	outKey := strings.TrimPrefix(key, "wall_connection_type_")
	switch value {
	case "short":
		return outKey, "low"
	default:
		return outKey, value
	}
}

// 还需考虑更早版本
func JavaTobedrockWallRelplacer(key string, value any) (string, any) {
	if key == "up" {
		return "wall_post_bit", value == `true`
	}
	switch value {
	case `false`:
		return "wall_connection_type_" + key, `none`
	case `true`, `low`:
		return "wall_connection_type_" + key, `short`
	default:
		return "wall_connection_type_" + key, value
	}
}
