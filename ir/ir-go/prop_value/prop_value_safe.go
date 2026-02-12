package prop_value

import (
	"fmt"
	"strconv"
	"strings"
)

type PropValue struct {
	v any
}

func NewPropValueFromString(v string) PropValue {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "\"") {
		if strings.HasSuffix(v, "\"") {
			v = v[1 : len(v)-1]
		} else {
			v = v[1:]
		}
	} else if strings.HasSuffix(v, "\"") {
		v = v[:len(v)-1]
	}
	return PropValue{
		v: v,
	}
}

func (v PropValue) IsString() bool {
	_, ok := v.v.(string)
	return ok
}

func (v PropValue) AsString() string {
	return v.v.(string)
}

func NewPropValueFromInt32(i int32) PropValue {
	return PropValue{
		v: i,
	}
}

func (v PropValue) IsInt32() bool {
	_, ok := v.v.(int32)
	return ok
}

func (v PropValue) AsInt32() int32 {
	return v.v.(int32)
}

func NewPropValueFromInt8(i int8) PropValue {
	return PropValue{
		v: i,
	}
}

func (v PropValue) IsInt8() bool {
	_, ok := v.v.(int8)
	return ok
}

func (v PropValue) AsInt8() int8 {
	return v.v.(int8)
}

func (v PropValue) String() string {
	if v.IsString() {
		return "\"" + v.AsString() + "\""
	}
	if v.IsInt32() {
		return fmt.Sprintf("%d", v.AsInt32())
	}
	if v.IsInt8() {
		return fmt.Sprintf("%db", v.AsInt8())
	}
	panic(fmt.Errorf("PropValue: unknown value %T:%v", v.v, v.v))
}

func (v PropValue) BedrockString() string {
	if v.IsString() {
		return "\"" + v.AsString() + "\""
	}
	if v.IsInt32() {
		return strconv.Itoa(int(v.AsInt32()))
	}
	if v.IsInt8() {
		if v.AsInt8() == 0 {
			return "false"
		} else {
			return "true"
		}
	}
	panic(fmt.Errorf("PropValue: unknown value %T:%v", v.v, v.v))
}
