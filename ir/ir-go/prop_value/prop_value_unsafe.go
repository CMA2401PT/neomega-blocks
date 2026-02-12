package prop_value

// import (
// 	"fmt"
// 	"strings"
// 	"unique"
// 	"unsafe"
// )

// type PropValue unique.Handle[string]

// func (v PropValue) toUintptr() uintptr {
// 	return *(*uintptr)(unsafe.Pointer(&v))
// }
// func boxInt32(i int32) PropValue {
// 	v := uintptr(uint32(i)) | int32Flag
// 	vp := &v
// 	va := (*unique.Handle[string])(unsafe.Pointer(vp))
// 	return PropValue(*va)
// }
// func boxInt8(i int8) PropValue {
// 	v := uintptr(uint8(i)) | int8Flag
// 	vp := &v
// 	va := (*unique.Handle[string])(unsafe.Pointer(vp))
// 	return PropValue(*va)
// }

// // highest bit
// const flag = ^((^uintptr(0)) >> 1)

// // 0 -> ptr to string
// // 10 -> int32
// // 110 -> byte
// const head1bit = flag
// const strFlag = 0
// const head2bit = flag | flag>>1
// const int32Flag = flag
// const head3bit = flag | flag>>1 | flag>>2
// const int8Flag = head2bit

// func NewPropValueFromString(v string) PropValue {
// 	v = strings.TrimSpace(v)
// 	if strings.HasPrefix(v, "\"") {
// 		if strings.HasSuffix(v, "\"") {
// 			v = v[1 : len(v)-1]
// 		} else {
// 			v = v[1:]
// 		}
// 	} else if strings.HasSuffix(v, "\"") {
// 		v = v[:len(v)-1]
// 	}
// 	return PropValue(unique.Make(v))
// }

// func (v PropValue) IsString() bool {
// 	return v.toUintptr()&head1bit == strFlag
// }

// func (v PropValue) AsString() string {
// 	return (unique.Handle[string])(v).Value()
// }

// func NewPropValueFromInt32(i int32) PropValue {
// 	return boxInt32(i)
// }

// func (v PropValue) IsInt32() bool {
// 	return v.toUintptr()&head2bit == int32Flag
// }

// func (v PropValue) AsInt32() int32 {
// 	return int32(uint32(v.toUintptr() ^ int32Flag))
// }

// func NewPropValueFromInt8(i int8) PropValue {
// 	return boxInt8(i)
// }

// func (v PropValue) IsInt8() bool {
// 	return v.toUintptr()&head3bit == int8Flag
// }

// func (v PropValue) AsInt8() int8 {
// 	return int8(uint8(v.toUintptr() ^ int8Flag))
// }

// func (v PropValue) String() string {
// 	if v.IsString() {
// 		return "\"" + v.AsString() + "\""
// 	}
// 	if v.IsInt32() {
// 		return fmt.Sprintf("%d", v.AsInt32())
// 	}
// 	if v.IsInt8() {
// 		return fmt.Sprintf("%db", v.AsInt8())
// 	}
// 	panic(fmt.Errorf("PropValue: unknown value %X", v.toUintptr()))
// }

// func (v PropValue) BedrockString() string {
// 	if v.IsString() {
// 		return "\"" + v.AsString() + "\""
// 	}
// 	if v.IsInt32() {
// 		return strconv.Itoa(int(v.AsInt32()))
// 	}
// 	if v.IsInt8() {
// 		if v.AsInt8() == 0 {
// 			return "false"
// 		} else {
// 			return "true"
// 		}
// 	}
// 	panic(fmt.Errorf("PropValue: unknown value %T:%v", v.v, v.v))
// }
