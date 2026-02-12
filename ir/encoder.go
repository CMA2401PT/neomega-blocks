package ir

import (
	"errors"
	encoder_bedrock_block_props "neomega-blocks/encode-decode/encoder-bedrock-block-props"
	nbt "neomega-nbt"
	seq_io "neomega-nbt/seq-io"
)

func (e *StateEncoder) WriteType(t nbt.TagType) error {
	if e.BedrockBlockPropsEncoder.W == nil {
		e.BedrockBlockPropsEncoder.W = seq_io.NewSliceGrowWriter(nil)
	}
	return e.BedrockBlockPropsEncoder.WriteType(t)
}

var errInvalidPropType = errors.New("invalid prop value type, can only be byte/int/string")
var errInvalidRootType = errors.New("root type must be compound")
var errInvalidRootTag = errors.New("root tag must be empty")
var errInvalidOperate = errors.New("invalid operate")
var errInvalidByte = errors.New("byte can only be 0/1")
var errInvalidInt = errors.New("int must be positive")

// 0 WriteType-> 1 WriteTag (empty) -> 2
// 2 if WriteType name(tag type=string) -> 3 WriteTag (name) -> 7 WriteString -> 2
// 2 if WriteType states(tag type=compound) -> 4 WriteTag (states) -> 8
// 2 if WriteEnd -> end
// 8 if WriteType / WriteTag / 10 WriteString/WriteInt/WirteByte / 8
// 8 if WriteEnd -> 2

func (e *BlockEncoder) WriteType(t nbt.TagType) error {
	switch e.aux {
	case 0:
		if t != nbt.TagCompound {
			return errInvalidRootType
		}
		e.aux = 1
		return nil
	case 2:
		switch t {
		case nbt.TagString:
			e.aux = 3
			return nil
		case nbt.TagCompound:
			e.aux = 4
			return nil
		default:
			return errInvalidOperate
		}
	case 8:
		return e.propsEnc.WriteType(t)
	}
	return errInvalidOperate
}

func (e *BlockEncoder) WriteTag(tag string) (err error) {
	switch e.aux {
	case 1:
		if tag != "" {
			return errInvalidRootTag
		}
		e.aux = 2
		return nil
	case 3:
		if tag == "name" {
			e.aux = 7
			return nil
		} else {
			return errInvalidOperate
		}
	case 4:
		if tag == "states" {
			e.aux = 8
			e.propsEnc = encoder_bedrock_block_props.BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]{
				W: seq_io.NewSliceGrowWriter(nil),
			}
			err = e.propsEnc.WriteType(nbt.TagCompound)
			if err != nil {
				return err
			}
			err = e.propsEnc.WriteTag("")
			if err != nil {
				return err
			}
			return nil
		} else {
			return errInvalidOperate
		}
	case 8:
		return e.propsEnc.WriteTag(tag)
	}
	return errInvalidOperate
}

func (e *BlockEncoder) WriteEmptyListMark() error {
	return errInvalidOperate
}

func (e *BlockEncoder) WriteLen(len uint32) error {
	return errInvalidOperate
}

func (e *BlockEncoder) WriteEnd() error {
	switch e.aux {
	case 2:
		props := IRStates(e.propsEnc.W.Bytes()).Tidy()
		e.block = (e.block + " " + string(props))
		return nil
	case 8:
		e.propsEnc.WriteEnd()
		e.aux = 2
		return nil
	}
	return errInvalidOperate
}

func (e *BlockEncoder) WriteInt8(i int8) (err error) {
	return e.propsEnc.WriteInt8(i)
}

func (e *BlockEncoder) WriteInt32(i int32) (err error) {
	return e.propsEnc.WriteInt32(i)
}

func (e *BlockEncoder) WriteString(v string) (err error) {
	switch e.aux {
	case 7:
		e.aux = 2
		e.block = v
		return nil
	case 8:
		return e.propsEnc.WriteString(v)
	default:
		return errInvalidOperate
	}
}

func (e *BlockEncoder) WriteInt16(int16) (err error) {
	return errInvalidOperate
}

func (e *BlockEncoder) WriteInt64(int64) (err error) {
	return errInvalidOperate
}

func (e *BlockEncoder) WriteFloat32(float32) (err error) {
	return errInvalidOperate
}

func (e *BlockEncoder) WriteFloat64(float64) (err error) {
	return errInvalidOperate
}
