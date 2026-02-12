package encoder_bedrock_block_props

import (
	"errors"
	"io"
	nbt "neomega-nbt"
	seq_io "neomega-nbt/seq-io"
	"strconv"
	"unsafe"
)

type BedrockBlockPropsEncoder[W seq_io.Writer] struct {
	W   W
	aux uint8
}

func NewEncoderFromIOWriter(w io.Writer) *BedrockBlockPropsEncoder[seq_io.OffsetWriter] {
	return &BedrockBlockPropsEncoder[seq_io.OffsetWriter]{W: seq_io.NewOffsetWriter(w)}
}

func NewEncoderToBytes(initBuf []uint8) (
	enc *BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter],
	getResult func() []uint8,
) {
	writer := seq_io.NewSliceGrowWriter(initBuf)
	return &BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]{W: writer}, func() []uint8 {
		return writer.Bytes()
	}
}

func NewEncoder[W seq_io.Writer](w W) *BedrockBlockPropsEncoder[W] {
	return &BedrockBlockPropsEncoder[W]{W: w}
}

var errInvalidPropType = errors.New("invalid prop value type, can only be byte/int/string")
var errInvalidRootType = errors.New("root type must be compound")
var errInvalidOperate = errors.New("invalid operate")
var errInvalidByte = errors.New("byte can only be 0/1")
var errInvalidInt = errors.New("int must be positive")

func (e *BedrockBlockPropsEncoder[W]) WriteType(t nbt.TagType) error {
	if e.aux == 0 {
		if t != nbt.TagCompound {
			return errInvalidRootType
		}
		e.aux = 1
		return nil
	}
	switch t {
	case nbt.TagByte, nbt.TagInt, nbt.TagString:
		return nil
	default:
		return errInvalidPropType
	}
}

func (e *BedrockBlockPropsEncoder[W]) WriteTag(tag string) (err error) {
	switch e.aux {
	case 1:
		e.aux = 2
		return e.W.WriteByte('[')
	case 2:
		e.aux = 3
	default:
		err = e.W.WriteByte(',')
		if err != nil {
			return
		}
	}

	err = e.W.WriteByte('"')
	if err != nil {
		return
	}
	err = e.W.Write(*(*[]byte)(unsafe.Pointer(&tag)))
	if err != nil {
		return
	}
	err = e.W.WriteByte('"')
	if err != nil {
		return
	}
	return e.W.WriteByte('=')
}

func (e BedrockBlockPropsEncoder[W]) WriteEmptyListMark() error {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteLen(len uint32) error {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteEnd() error {
	return e.W.WriteByte(']')
}

func (e BedrockBlockPropsEncoder[W]) WriteInt8(i int8) (err error) {
	switch i {
	case 0:
		err = e.W.WriteByte('f')
		if err != nil {
			return
		}
		err = e.W.WriteByte('a')
		if err != nil {
			return
		}
		err = e.W.WriteByte('l')
		if err != nil {
			return
		}
		err = e.W.WriteByte('s')
		if err != nil {
			return
		}
		return e.W.WriteByte('e')
	case 1:
		err = e.W.WriteByte('t')
		if err != nil {
			return
		}
		err = e.W.WriteByte('r')
		if err != nil {
			return
		}
		err = e.W.WriteByte('u')
		if err != nil {
			return
		}
		return e.W.WriteByte('e')
	default:
		return errInvalidByte
	}
}

func (e BedrockBlockPropsEncoder[W]) WriteInt16(int16) (err error) {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteInt32(i int32) (err error) {
	if i < 0 {
		return errInvalidInt
	}
	if i == 0 {
		return e.W.WriteByte('0')
	}
	digitCount := 0
	v := i
	for v > 0 {
		v = v / 10
		digitCount += 1
	}
	for p := range digitCount {
		v := i
		for j := p + 1; j < digitCount; j++ {
			v /= 10
		}
		v = v % 10
		err = e.W.WriteByte(uint8(int32('0') + v))
		if err != nil {
			return
		}
	}
	return nil
}

func (e BedrockBlockPropsEncoder[W]) WriteInt64(int64) (err error) {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteFloat32(float32) (err error) {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteFloat64(float64) (err error) {
	return errInvalidOperate
}

func (e BedrockBlockPropsEncoder[W]) WriteString(v string) (err error) {
	v = strconv.Quote(v)
	return e.W.Write(*(*[]byte)(unsafe.Pointer(&v)))
}
