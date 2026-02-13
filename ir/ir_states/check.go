package ir_states

import (
	"fmt"
	"strconv"
	"strings"
)

func irStateValueToString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		trimmed := strings.TrimSpace(x)
		lower := strings.ToLower(trimmed)
		if lower == "true" {
			return "1", true
		}
		if lower == "false" {
			return "0", true
		}
		if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return strconv.FormatInt(n, 10), true
		}
		return x, true
	case bool:
		if x {
			return "1", true
		}
		return "0", true
	case int8:
		if x == 0 {
			return "0", true
		}
		if x == 1 {
			return "1", true
		}
		return strconv.FormatInt(int64(x), 10), true
	case int16:
		return strconv.FormatInt(int64(x), 10), true
	case int32:
		return strconv.FormatInt(int64(x), 10), true
	case int64:
		return strconv.FormatInt(x, 10), true
	case int:
		return strconv.Itoa(x), true
	case uint8:
		return strconv.FormatUint(uint64(x), 10), true
	case uint16:
		return strconv.FormatUint(uint64(x), 10), true
	case uint32:
		return strconv.FormatUint(uint64(x), 10), true
	case uint64:
		return strconv.FormatUint(x, 10), true
	case uint:
		return strconv.FormatUint(uint64(x), 10), true
	default:
		return fmt.Sprint(v), false
	}
}

func CheckKey(s string, key string, val any) bool {
	got, found := Get(s, key)
	if !found {
		return false
	}

	switch gv := got.(type) {
	case string:
		if vv, ok := val.(string); ok {
			return gv == vv
		}
	case int32:
		if vv, ok := val.(int32); ok {
			return gv == vv
		}
	case int8:
		if vv, ok := val.(int8); ok {
			return gv == vv
		}
	}

	gs, ok1 := irStateValueToString(got)
	vs, ok2 := irStateValueToString(val)
	if !ok1 || !ok2 {
		return false
	}
	return gs == vs
}
