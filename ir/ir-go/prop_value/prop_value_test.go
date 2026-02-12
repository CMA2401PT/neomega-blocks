package prop_value

import (
	"strconv"
	"strings"
	"testing"
)

// func TestFlag(t *testing.T) {
// 	s := fmt.Sprintf("%b", flag)
// 	if strings.Count(s, "1") != 1 {
// 		t.Fail()
// 	}
// 	if strings.Count(s, "0") != 63 && strings.Count(s, "0") != 31 {
// 		t.Fail()
// 	}
// }

func TestString(t *testing.T) {
	pv := NewPropValueFromString("minecraft")
	if !pv.IsString() {
		t.Fail()
	}
	if pv.AsString() != "minecraft" {
		t.Fail()
	}
}

func TestInt32(t *testing.T) {
	check := func(v int32) {
		pv := NewPropValueFromInt32(v)
		if !pv.IsInt32() {
			t.Errorf("check %v int32 fail", v)
		}
		if pv.AsInt32() != v {
			t.Errorf("check %v int32 value fail, get: %v", v, pv.AsInt32())
		}
	}

	check(0)
	check(-1)
	check(1)
	check(1 << 20)
	check(-(1 << 20))
}

func TestInt8(t *testing.T) {
	check := func(v int8) {
		pv := NewPropValueFromInt8(v)
		if !pv.IsInt8() {
			t.Errorf("check %v int8 fail", v)
		}
		if pv.AsInt8() != v {
			t.Errorf("check %v int8 value fail, get: %v", v, pv.AsInt32())
		}
	}

	check(0)
	check(-1)
	check(1)
	check(1 << 6)
	check(-(1 << 6))
}

func TestStringPropValue(t *testing.T) {
	check := func(inStr string, expectStr string, expectInt int32, expectByte int8) {
		pv := NewPropValueFromString(inStr)
		if pv.String() != expectStr {
			t.Errorf("prop value %s should be %s, but got %s", inStr, expectStr, pv.String())
		}
		if !FuzzyEqual(pv, NewPropValueFromInt8(expectByte)) {
			t.Errorf("prop value %s should be int:%d byte:%d", inStr, expectInt, expectByte)
		}
		if !FuzzyEqual(pv, NewPropValueFromInt32(expectInt)) {
			t.Errorf("prop value %s should be int:%d byte:%d", inStr, expectInt, expectByte)
		}
	}
	check("true", "\"true\"", 1, 1)
	check("false", "\"false\"", 0, 0)
	check("True", "\"True\"", 1, 1)
	check("False", "\"False\"", 0, 0)
	check("TRUE", "\"TRUE\"", 1, 1)
	check("FALSE", "\"FALSE\"", 0, 0)
	check("0b", "\"0b\"", 0, 0)
	check("1b", "\"1b\"", 1, 1)
	check("2b", "\"2b\"", 2, 2)
	check("0", "\"0\"", 0, 0)
	check("1", "\"1\"", 1, 1)
	check("2", "\"2\"", 2, 2)
	check("0L", "\"0L\"", 0, 0)
	check("1L", "\"1L\"", 1, 1)
	check("2L", "\"2L\"", 2, 2)

	checkStr := func(inStr string, expectStr string) {
		pv := NewPropValueFromString(inStr)
		if pv.String() != expectStr {
			t.Errorf("prop value %s should be %s, but got %s", inStr, expectStr, pv.String())
		}
	}
	checkStr("a", "\"a\"")
	checkStr("b", "\"b\"")
	checkStr("c", "\"c\"")
	checkStr("stone", "\"stone\"")
	checkStr("\"stone\"", "\"stone\"")
	checkStr("minecraft:stone", "\"minecraft:stone\"")
	checkStr("\"minecraft:stone\"", "\"minecraft:stone\"")
}

func TestPropValueString(t *testing.T) {
	expectStr := func(inStr string, expectStr string) {
		v := NewPropValueFromString(inStr)
		if v.String() != expectStr {
			t.Errorf("expect %v, get %v", expectStr, v.String())
		}
	}
	expectStr("a", `"a"`)
	expectStr(`"a"`, `"a"`)
	expectStr("stone", `"stone"`)
	expectStr(`"stone"`, `"stone"`)
	expectStr(`"123"`, `"123"`)
	expectStr(`"123L"`, `"123L"`)
	expectStr(`"123l"`, `"123l"`)
	expectStr(`"123I"`, `"123I"`)
	expectStr(`"123i"`, `"123i"`)
	expectStr(`"123S"`, `"123S"`)
	expectStr(`"123s"`, `"123s"`)
	expectStr(`"123B"`, `"123B"`)
	expectStr(`"123b"`, `"123b"`)
	expectStr(`"true"`, `"true"`)
	expectStr(`"false"`, `"false"`)
	expectStr(`"True"`, `"True"`)
	expectStr(`"False"`, `"False"`)
	expectStr(`"TRUE"`, `"TRUE"`)
	expectStr(`"FALSE"`, `"FALSE"`)

	expectSnbtStr := func(inStr string, expectStr string) {
		v := NewPropValueFromSNBT(inStr)
		if v.String() != expectStr {
			t.Errorf("expect %v, get %v", expectStr, v.String())
		}
	}
	expectSnbtStr(`1b`, `1b`)
	expectSnbtStr(`"1b"`, `"1b"`)
	expectSnbtStr(`0b`, `0b`)
	expectSnbtStr(`"0b"`, `"0b"`)
	expectSnbtStr(`true`, `1b`)
	expectSnbtStr(`false`, `0b`)
	expectSnbtStr(`123`, `123`)
	expectSnbtStr("stone", `"stone"`)
	expectSnbtStr(`"stone"`, `"stone"`)
}

// test make value from snbt, check type
// "true" = Byte(1) -> 1b
// "\"true\"" = String("true") -> "\"true\""
// "false" = Byte(0) -> 0b
// "\"false\"" = String("false") -> "\"false\""
// "0b" = Byte(0) -> 0b
// "\"0b\"" = String("0b") -> "\"0b\""
// "1b" = Byte(1) -> 1b
// "\"1b\"" = String("1b") -> "\"1b\""
// "0" = Byte(0) -> 0
// "\"0\"" = String("0") -> "\"0\""
// "1" = Byte(1) -> 1
// "\"1\"" = String("1") -> "\"1\""
// "0L" = String("0L") -> "\"0L\""
// "\"0L\"" = String("0L") -> "\"0L\""
// "1L" = String("1L") -> "\"1L\""
// "\"1L\"" = String("1L") -> "\"1L\""
// "stone" = String("stone") -> "\"stone\""
// "\"stone\"" = String("stone") -> "\"stone\""
// "minecraft:stone" = String("minecraft:stone") -> "\"minecraft:stone\""
// "\"minecraft:stone\"" = String("minecraft:stone") -> "\"minecraft:stone\""
func TestMakePropValueFromSNBT(t *testing.T) {
	check := func(inStr string, expectStr string, expectValue PropValue) {
		pv := NewPropValueFromSNBT(inStr)
		if pv != (expectValue) {
			t.Errorf("prop value %s should be value %s, but got %s", inStr, expectValue, pv.String())
		}
		if pv.String() != expectStr {
			t.Errorf("prop value %s should be %s, but got %s", inStr, expectStr, pv.String())
		}
	}
	String := NewPropValueFromString
	Byte := NewPropValueFromInt8
	Int := NewPropValueFromInt32
	check("true", "1b", Byte(1))
	check("\"true\"", "\"true\"", String("true"))
	check("false", "0b", Byte(0))
	check("\"false\"", "\"false\"", String("false"))
	check("0b", "0b", Byte(0))
	check("\"0b\"", "\"0b\"", String("0b"))
	check("1b", "1b", Byte(1))
	check("\"1b\"", "\"1b\"", String("1b"))
	// check("2b", "2b", Byte(2))
	// check("\"2b\"", "\"2b\"", String("2b"))
	check("0", "0", Int(0))
	check("\"0\"", "\"0\"", String("0"))
	check("1", "1", Int(1))
	check("\"1\"", "\"1\"", String("1"))
	check("2", "2", Int(2))
	check("\"2\"", "\"2\"", String("2"))
	check("0L", "\"0L\"", String("0L"))
	check("\"0L\"", "\"0L\"", String("0L"))
	check("1L", "\"1L\"", String("1L"))
	check("\"1L\"", "\"1L\"", String("1L"))
	check("stone", "\"stone\"", String("stone"))
	check("\"stone\"", "\"stone\"", String("stone"))
	check("minecraft:stone", "\"minecraft:stone\"", String("minecraft:stone"))
	check("\"minecraft:stone\"", "\"minecraft:stone\"", String("minecraft:stone"))
}

// test make value from snbt, test bedrock string
// "true" = Byte(1) -> "true"
// "\"true\"" = String("true") -> "\"true\""
// "false" = Byte(0) -> "false"
// "\"false\"" = String("false") -> "\"false\""
// "0b" = Byte(0) -> "false"
// "\"0b\"" = String("0b") -> "\"0b\""
// "1b" = Byte(1) -> "true"
// "\"1b\"" = String("1b") -> "\"1b\""
// "0" = Int(0) -> "false"
// "\"0\"" = String("0") -> "\"0\""
// "1" = Int(1) -> "true"
// "\"1\"" = String("1") -> "\"1\""
// "0L" = String("0L") -> "\"0L\""
// "\"0L\"" = String("0L") -> "\"0L\""
// "1L" = String("1L") -> "\"1L\""
// "\"1L\"" = String("1L") -> "\"1L\""
// "stone" = String("stone") -> "\"stone\""
// "\"stone\"" = String("stone") -> "\"stone\""
// "minecraft:stone" = String("minecraft:stone") -> "\"minecraft:stone\""
// "\"minecraft:stone\"" = String("minecraft:stone") -> "\"minecraft:stone\""
// func TestMakePropValueFromSNBTBedrock(t *testing.T) {
// 	check := func(inStr string, expectStr string, expectValue PropValue) {
// 		pv := NewPropValueFromSNBT(inStr)
// 		if pv != (expectValue) {
// 			t.Errorf("prop value %s should be value %s, but got %s", inStr, expectValue, pv.String())
// 		}
// 		if pv.BedrockString() != expectStr {
// 			t.Errorf("prop value %s should be %s, but got %s", inStr, expectStr, pv.BedrockString())
// 		}
// 	}
// 	String := NewPropValueFromString
// 	Byte := NewPropValueFromInt8
// 	Int := NewPropValueFromInt32
// 	check("true", "true", Byte(1))
// 	check("\"true\"", "\"true\"", String("true"))
// 	check("false", "false", Byte(0))
// 	check("\"false\"", "\"false\"", String("false"))
// 	check("0b", "false", Byte(0))
// 	check("\"0b\"", "\"0b\"", String("0b"))
// 	check("1b", "true", Byte(1))
// 	check("\"1b\"", "\"1b\"", String("1b"))
// 	check("2b", "true", Byte(2))
// 	check("\"2b\"", "\"2b\"", String("2b"))
// 	check("0", "0", Int(0))
// 	check("\"0\"", "\"0\"", String("0"))
// 	check("1", "1", Int(1))
// 	check("\"1\"", "\"1\"", String("1"))
// 	check("2", "2", Int(2))
// 	check("\"2\"", "\"2\"", String("2"))
// 	check("0L", "\"0L\"", String("0L"))
// 	check("\"0L\"", "\"0L\"", String("0L"))
// 	check("1L", "\"1L\"", String("1L"))
// 	check("\"1L\"", "\"1L\"", String("1L"))
// 	check("stone", "\"stone\"", String("stone"))
// 	check("\"stone\"", "\"stone\"", String("stone"))
// 	check("minecraft:stone", "\"minecraft:stone\"", String("minecraft:stone"))
// 	check("\"minecraft:stone\"", "\"minecraft:stone\"", String("minecraft:stone"))
// }

func NewPropValueFromSNBT(v string) PropValue {
	v = strings.TrimSpace(v)
	if v == "true" || v == "1b" {
		return NewPropValueFromInt8(1)
	}
	if v == "false" || v == "" || v == "0b" {
		return NewPropValueFromInt8(0)
	}
	if strings.HasPrefix(v, "\"") {
		return NewPropValueFromString(v)
	}
	if strings.HasSuffix(v, "b") {
		val, err := strconv.ParseUint(v[:len(v)-1], 10, 16)
		if err == nil {
			return NewPropValueFromInt8(int8(val))
		}

	}
	val, err := strconv.ParseUint(v, 10, 16)
	if err == nil {
		return NewPropValueFromInt32(int32(val))
	}
	return NewPropValueFromString(v)
}
