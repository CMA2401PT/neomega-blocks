package ir_go

// import (
// 	"errors"
// 	"neomega-blocks/ir/ir-go/prop_value"
// 	nbt "neomega-nbt"
// 	decoder_go_object "neomega-nbt/decoder-go-object"
// 	decoder_shim "neomega-nbt/decoder-shim"
// )

// type Encoder struct {
// 	// 0 WriteType-> 1 WriteTag (empty) -> 2
// 	// 2 if WriteType name(tag type=string) -> 3 WriteTag (name) -> 7 WriteString -> 2
// 	// 2 if WriteType states(tag type=compound) -> 4 WriteTag (states) -> 8
// 	// 2 if WriteEnd -> end
// 	// 8 if WriteType -> 9 WriteTag -> 10 WriteString/WriteInt/WirteByte -> 8
// 	// 8 if WriteEnd -> 2
// 	aux   uint8
// 	block *IRGoBlock
// }

// var errInvalidPropType = errors.New("invalid prop value type, can only be byte/int/string")
// var errInvalidRootType = errors.New("root type must be compound")
// var errInvalidRootTag = errors.New("root tag must be empty")
// var errInvalidOperate = errors.New("invalid operate")
// var errInvalidByte = errors.New("byte can only be 0/1")
// var errInvalidInt = errors.New("int must be positive")

// func (e *Encoder) Block() *IRGoBlock {
// 	return e.block
// }
// func (e *Encoder) WriteType(t nbt.TagType) error {
// 	switch e.aux {
// 	case 0:
// 		if t != nbt.TagCompound {
// 			return errInvalidRootType
// 		}
// 		e.block = &IRGoBlock{}
// 		e.aux = 1
// 		return nil
// 	case 2:
// 		switch t {
// 		case nbt.TagString:
// 			e.aux = 3
// 			return nil
// 		case nbt.TagCompound:
// 			e.aux = 4
// 			return nil
// 		default:
// 			return errInvalidOperate
// 		}
// 	case 8:
// 		switch t {
// 		case nbt.TagByte, nbt.TagInt, nbt.TagString:
// 			e.aux = 9
// 			return nil
// 		default:
// 			return errInvalidPropType
// 		}
// 	}
// 	return errInvalidOperate
// }

// func (e *Encoder) WriteTag(tag string) (err error) {
// 	switch e.aux {
// 	case 1:
// 		if tag != "" {
// 			return errInvalidRootTag
// 		}
// 		e.aux = 2
// 		return nil
// 	case 3:
// 		if tag == "name" {
// 			e.aux = 7
// 			return nil
// 		} else {
// 			return errInvalidOperate
// 		}
// 	case 4:
// 		if tag == "states" {
// 			e.aux = 8
// 			return nil
// 		} else {
// 			return errInvalidOperate
// 		}
// 	case 9:
// 		e.aux = 10
// 		if e.block.props == nil {
// 			e.block.props = make([]IRGoProp, 0)
// 		}
// 		e.block.props = append(e.block.props, IRGoProp{Key: tag})
// 		return nil
// 	}
// 	return errInvalidOperate
// }

// func (e Encoder) WriteEmptyListMark() error {
// 	return errInvalidOperate
// }

// func (e Encoder) WriteLen(len uint32) error {
// 	return errInvalidOperate
// }

// func (e *Encoder) WriteEnd() error {
// 	switch e.aux {
// 	case 2:
// 		e.block.sort()
// 		return nil
// 	case 8:
// 		e.aux = 2
// 		return nil
// 	}
// 	return errInvalidOperate
// }

// func (e *Encoder) WriteInt8(i int8) (err error) {
// 	if e.aux != 10 {
// 		return errInvalidOperate
// 	}
// 	e.aux = 8
// 	p := e.block.props[len(e.block.props)-1]
// 	p.Value = prop_value.NewPropValueFromInt8(i)
// 	e.block.props[len(e.block.props)-1] = p
// 	return nil
// }

// func (e Encoder) WriteInt16(int16) (err error) {
// 	return errInvalidOperate
// }

// func (e *Encoder) WriteInt32(i int32) (err error) {
// 	if e.aux != 10 {
// 		return errInvalidOperate
// 	}
// 	e.aux = 8
// 	p := e.block.props[len(e.block.props)-1]
// 	p.Value = prop_value.NewPropValueFromInt32(i)
// 	e.block.props[len(e.block.props)-1] = p
// 	return nil
// }

// func (e Encoder) WriteInt64(int64) (err error) {
// 	return errInvalidOperate
// }

// func (e Encoder) WriteFloat32(float32) (err error) {
// 	return errInvalidOperate
// }

// func (e Encoder) WriteFloat64(float64) (err error) {
// 	return errInvalidOperate
// }

// func (e *Encoder) WriteString(v string) (err error) {
// 	switch e.aux {
// 	case 7:
// 		e.aux = 2
// 		e.block.nameWithNameSpace = v
// 		return nil
// 	case 10:
// 		e.aux = 8
// 		p := e.block.props[len(e.block.props)-1]
// 		p.Value = prop_value.NewPropValueFromString(v)
// 		e.block.props[len(e.block.props)-1] = p
// 		return nil
// 	default:
// 		return errInvalidOperate
// 	}
// }

// func NewIRBlockFromGo(nameWithNamespace string, props map[string]any) (*IRGoBlock, error) {
// 	if props == nil {
// 		props = map[string]any{}
// 	}
// 	enc := &Encoder{}
// 	err := decoder_go_object.HighLevelDecodeTo(
// 		map[string]any{
// 			"name":   nameWithNamespace,
// 			"states": props,
// 		},
// 		decoder_shim.NewDecoderShim(enc),
// 		decoder_go_object.DecodeOption{
// 			SortMapKey: false,
// 		},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return enc.block, nil
// }
