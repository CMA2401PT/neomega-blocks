package ir

import (
	decoder_bedrock_block_props "neomega-blocks/encode-decode/decoder-bedrock-block-props"
	encoder_bedrock_block_props "neomega-blocks/encode-decode/encoder-bedrock-block-props"
	encoder_block_hash "neomega-blocks/encode-decode/encoder-block-hash"
	"neomega-blocks/ir/ir_states"
	nbt "neomega-nbt"
	decoder_go_object "neomega-nbt/decoder-go-object"
	decoder_shim "neomega-nbt/decoder-shim"
	seq_io "neomega-nbt/seq-io"
	"strings"
)

// 结构类似 minecraft:block [], minecraft:block ["key1"="string","key2"=true,"key3"=false,"key4"=123]
type IRBlock string

func (b IRBlock) String() string {
	return string(b)
}
func (b IRBlock) NameWithNameSpace() string {
	idx := strings.IndexByte(string(b), ' ')
	if idx == -1 {
		return string(b)
	}
	return string(b)[:idx]
}

func (b IRBlock) ShortName() string {
	idxCol := strings.IndexByte(string(b), ':')
	idxSpc := strings.IndexByte(string(b), ' ')
	if idxSpc == -1 {
		if idxCol == -1 {
			return string(b)
		}
		return string(b)[idxCol+1:]
	} else {
		if idxCol == -1 {
			return string(b)[:idxSpc]
		}
		return string(b)[idxCol+1 : idxSpc]
	}
}

func (b IRBlock) NameSpace() string {
	idxCol := strings.IndexByte(string(b), ':')
	if idxCol == -1 {
		return ""
	} else {
		return string(b)[:idxCol]
	}
}

func (b IRBlock) States() IRStates {
	idxCol := strings.IndexByte(string(b), ' ')
	if idxCol == -1 {
		return "[]"
	} else {
		return IRStates(string(b)[idxCol+1:])
	}
}

func (b IRBlock) IsTidy() bool {
	s := string(b)
	idxSp := strings.IndexByte(s, ' ')
	if idxSp <= 0 || idxSp >= len(s)-1 {
		return false
	}
	// exactly one separator space between name and states
	if s[idxSp+1] == ' ' {
		return false
	}
	name := s[:idxSp]
	states := s[idxSp+1:]

	idxCol := strings.IndexByte(name, ':')
	if idxCol <= 0 || idxCol >= len(name)-1 {
		return false
	}
	if strings.LastIndexByte(name, ':') != idxCol {
		return false
	}
	return IRStates(states).IsTidy()
}

func (b IRBlock) Tidy() IRBlock {
	raw := strings.TrimSpace(string(b))
	if raw == "" {
		return b
	}

	name := raw
	statesRaw := "[]"
	if idxSp := strings.IndexByte(raw, ' '); idxSp >= 0 {
		name = raw[:idxSp]
		statesRaw = strings.TrimSpace(raw[idxSp+1:])
		if statesRaw == "" {
			statesRaw = "[]"
		}
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return b
	}

	if strings.IndexByte(name, ':') == -1 {
		name = "minecraft:" + name
	}

	states := IRStates(statesRaw).Tidy()
	if !states.IsTidy() {
		states = "[]"
	}
	return IRBlock(name + " " + string(states))
}

func (b IRBlock) IsValid() bool {
	raw := strings.TrimSpace(string(b))
	if raw == "" {
		return false
	}

	name := raw
	statesRaw := "[]"
	if idxSp := strings.IndexByte(raw, ' '); idxSp >= 0 {
		name = strings.TrimSpace(raw[:idxSp])
		statesRaw = strings.TrimSpace(raw[idxSp+1:])
		if statesRaw == "" {
			statesRaw = "[]"
		}
	}
	if name == "" {
		return false
	}
	if strings.ContainsAny(name, `[]",=`) {
		return false
	}
	if idxCol := strings.IndexByte(name, ':'); idxCol >= 0 {
		if strings.LastIndexByte(name, ':') != idxCol {
			return false
		}
		if idxCol == len(name)-1 {
			return false
		}
	}
	return IRStates(statesRaw).IsValid()
}

func (b IRBlock) Hash() uint32 {
	if b.NameWithNameSpace() == "minecraft:unknown" {
		return ^uint32(1)
	}
	enc := encoder_block_hash.NewEncoder()
	decodeToHL(b, decoder_shim.NewDecoderShim(enc))
	return enc.Hash()
}

type IRStates string

func (s IRStates) IsTidy() bool {
	return ir_states.IsTidy((string(s)))
}

func (s IRStates) IsValid() bool {
	return ir_states.IsValid((string(s)))
}

func (s IRStates) Tidy() IRStates {
	return IRStates(ir_states.Tidy(string(s)))
}

func (s IRStates) KeyValueStrings(yield func(string, string) bool) {
	ir_states.KeyValueStrings((string(s)))(yield)
}

func (s IRStates) KeyValues(yield func(string, any) bool) {
	ir_states.KeyValues((string(s)))(yield)
}

func (s IRStates) Get(key string) (value any, found bool) {
	return ir_states.Get(string(s), key)
}

func (s IRStates) Set(key string, value any) IRStates {
	return IRStates(ir_states.Set(string(s), key, value))
}

func (s IRStates) Delete(key string) IRStates {
	return IRStates(ir_states.Delete(string(s), key))
}

// 将所有的 value 全部转为 string 类型，
// 若原本为 true/false 则转为 "0"/"1"
// 若原本为 int，则转为 类似 "123" 的形式
func (s IRStates) CastToString() IRStates {
	return IRStates(ir_states.CastToString(string(s)))
}

type CompareResult struct {
	// 若两个 states 完全相同，返回这个
	ExactlySame bool
	// 相同（value 映射为String之后）的属性个数
	Same int
	// 都存在且不同的属性个数
	Different int
	// s1 缺失的属性个数(s2存在)
	Missing int
	// s1 多余的属性个数(s2不存在)
	Extra int
}

func (s1 IRStates) FuzzyCompare(s2 IRStates) CompareResult {
	if s1 == s2 {
		return CompareResult{ExactlySame: true}
	}
	s1u := s1.CastToString()
	s2u := s2.CastToString()
	s, d, m, e := ir_states.FuzzyCompare(string(s1u), string(s2u))
	return CompareResult{ExactlySame: false, Same: s, Different: d, Missing: m, Extra: e}
}

// decoders
func (s IRStates) DecodeToHL(enc nbt.EncoderHighLevel) error {
	return decoder_bedrock_block_props.Decode(s, enc, true)
}

func (s IRStates) DecodeToLL(enc nbt.Encoder) error {
	return decoder_bedrock_block_props.Decode(s, decoder_shim.NewDecoderShim(enc), true)
}

func (b IRBlock) DecodeToHL(encoder nbt.EncoderHighLevel) (err error) {
	return decodeToHL(b, encoder)
}

func (b IRBlock) DecodeToLL(encoder nbt.Encoder) (err error) {
	return decodeToHL(b, decoder_shim.NewDecoderShim(encoder))
}

// encoders
type StateEncoder struct {
	encoder_bedrock_block_props.BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]
}

func (enc *StateEncoder) States() IRStates {
	return IRStates(enc.BedrockBlockPropsEncoder.W.Bytes()).Tidy()
}

func NewIRStateEncoder() *StateEncoder {
	return &StateEncoder{}
}

type BlockEncoder struct {
	aux      uint8
	propsEnc encoder_bedrock_block_props.BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]
	block    string
}

func NewIRBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

func (e *BlockEncoder) Block() IRBlock {
	return IRBlock(e.block)
}

func NewIRBlockFromGo(nameWithNamespace string, props map[string]any) (IRBlock, error) {
	if props == nil {
		props = map[string]any{}
	}
	enc := &BlockEncoder{}
	err := decoder_go_object.HighLevelDecodeTo(
		map[string]any{
			"name":   nameWithNamespace,
			"states": props,
		},
		decoder_shim.NewDecoderShim(enc),
		decoder_go_object.DecodeOption{
			SortMapKey: false,
		},
	)
	if err != nil {
		return "", err
	}
	return enc.Block(), nil
}
