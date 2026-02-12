package decoder_bedrock_block_props

import (
	"errors"
	"fmt"
	"neomega-blocks/ir/ir_states"
	nbt "neomega-nbt"
)

var errNoLeftBracket = errors.New("missing left [")
var errNoRightBracket = errors.New("missing right ]")
var errNoBeingingKey = errors.New("missing \" at key begining")
var errNoEndKey = errors.New("missing \" at key end")
var errNoIdentifier = errors.New("missing = between key and value")
var errMissingValue = errors.New("missing value")
var errUnknowValueType = errors.New("value can only be true/false/positive int/string")

// ["byteKey"=true/false,"intKey"=0/1/2/3...,"stringKey"="string"]
func Decode[Seq ~[]uint8 | ~string, E nbt.EncoderHighLevel](props Seq, enc E, withRootTag bool) (err error) {
	if withRootTag {
		err = enc.WriteRootTag("", nbt.ComplexCompound)
		if err != nil {
			return err
		}
	}

	err = enc.BeginComplexCompound()
	if err != nil {
		return err
	}
	isFirst := true
	for key, value := range ir_states.KeyValues(string(props)) {
		if !isFirst {
			err = enc.NextCompoundElement()
			if err != nil {
				return err
			}
		} else {
			isFirst = false
		}
		switch vt := value.(type) {
		case int8:
			err = enc.WriteCompoundTag(key, nbt.TrivialByte)
			if err != nil {
				return err
			}
			err = enc.WriteInt8(vt, nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		case int32:
			err = enc.WriteCompoundTag(key, nbt.TrivialInt)
			if err != nil {
				return err
			}
			err = enc.WriteInt32(vt, nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		case string:
			err = enc.WriteCompoundTag(key, nbt.TrivialString)
			if err != nil {
				return err
			}
			err = enc.WriteString(vt, nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		default:
			panic(fmt.Errorf("unknown element type: %T:%v", value, value))
		}

	}
	return enc.EndComplexCompound()
}
