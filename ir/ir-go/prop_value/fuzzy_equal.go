package prop_value

import (
	"fmt"
	"strconv"
)

// check is special string and return number (cast to int32)
// e.g. 1L, 1b, True, False
func fuzzyAsNumber(v string) (pv int32, ok bool) {
	val, err := strconv.ParseInt(v, 10, 31)
	if err == nil {
		return int32(val), true
	}
	switch v {
	case "false":
		return 0, true
	case "true":
		return 1, true
	case "False":
		return 0, true
	case "True":
		return 1, true
	case "FALSE":
		return 0, true
	case "TRUE":
		return 1, true
	}
	if len(v) > 1 {
		val, err := strconv.ParseUint(v[:len(v)-1], 10, 31)
		if err == nil {
			return int32(val), true
		}
	}
	return 0, false
}

func fuzzyCastInt32(v PropValue) (pv int32, ok bool) {
	if v.IsString() {
		return fuzzyAsNumber(v.AsString())
	}
	if v.IsInt32() {
		return v.AsInt32(), true
	}
	if v.IsInt8() {
		return int32(v.AsInt8()), true
	}
	panic(fmt.Errorf("PropValue: unknown value %v", v.String()))
}

func FuzzyEqual(v1, v2 PropValue) bool {
	if v1 == v2 {
		return true
	}
	var v1i, v2i int32
	var ok bool
	if v1i, ok = fuzzyCastInt32(v1); !ok {
		return false
	}
	if v2i, ok = fuzzyCastInt32(v2); !ok {
		return false
	}
	return v1i == v2i
}

func (v PropValue) FuzzyEqual(v2 PropValue) bool {
	return FuzzyEqual(v, v2)
}
