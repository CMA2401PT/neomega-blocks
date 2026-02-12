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

// IRBlock 是方块的文本中间表示。
// 常见形式：
//   - minecraft:block []
//   - minecraft:block ["key1"="string","key2"=true,"key3"=false,"key4"=123]
type IRBlock string

// String 返回 IRBlock 的原始文本。
func (b IRBlock) String() string {
	return string(b)
}

// NameWithNameSpace 返回第一个空格前的方块名部分。
// 如果没有 states，则返回整个字符串。
func (b IRBlock) NameWithNameSpace() string {
	idx := strings.IndexByte(string(b), ' ')
	if idx == -1 {
		return string(b)
	}
	return string(b)[:idx]
}

// ShortName 返回去掉命名空间和 states 后的短方块名。
// 例如：minecraft:stone ["a"=1] -> stone。
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

// NameSpace 返回 ':' 前的命名空间部分。
// 如果不存在命名空间，返回空字符串。
func (b IRBlock) NameSpace() string {
	idxCol := strings.IndexByte(string(b), ':')
	if idxCol == -1 {
		return ""
	} else {
		return string(b)[:idxCol]
	}
}

// States 返回第一个空格后的 states 部分。
// 如果 states 缺失，返回 []。
func (b IRBlock) States() IRStates {
	idxCol := strings.IndexByte(string(b), ' ')
	if idxCol == -1 {
		return "[]"
	} else {
		return IRStates(string(b)[idxCol+1:])
	}
}

// IsTidy 判断方块文本是否已处于规范整洁形式：
// 必须包含命名空间，name 与 states 之间恰好一个空格，且 states 本身是 tidy 的。
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

// Tidy 将方块文本规范化为标准形式。
// 当命名空间缺失时会补上默认 "minecraft:"，并规范化 states。
// 如果 states 无法规范化，回退为 []。
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

// IsValid 判断方块文本在语法上是否有效。
// 同时接受带命名空间和不带命名空间的名称。
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

// Hash 返回该方块对应的 Bedrock 网络哈希值。
// 特殊方块 minecraft:unknown 会映射为 ^uint32(1)。
func (b IRBlock) Hash() uint32 {
	if b.NameWithNameSpace() == "minecraft:unknown" {
		return ^uint32(1)
	}
	enc := encoder_block_hash.NewEncoder()
	decodeToHL(b, decoder_shim.NewDecoderShim(enc))
	return enc.Hash()
}

// IRStates 是方块 states 的文本表示。
// 例如：["facing_direction"=2,"open_bit"=false]
type IRStates string

// IsTidy 判断 states 是否已处于规范整洁形式。
func (s IRStates) IsTidy() bool {
	return ir_states.IsTidy((string(s)))
}

// IsValid 判断 states 在语法上是否有效。
func (s IRStates) IsValid() bool {
	return ir_states.IsValid((string(s)))
}

// Tidy 返回规范化后的 states 文本（空白与顺序都会标准化）。
func (s IRStates) Tidy() IRStates {
	return IRStates(ir_states.Tidy(string(s)))
}

// Canon 从 std 中增改 keys
// std 中包括允许的所有 keys 的默认值
// 当 s 中缺少 std 中的某个 key 时，从 std 中的 key 设置 s 的值
// 当 s 中包括 std 中不存在的 key 时，从 s 中移除这个 key
func (s IRStates) Canon(std IRStates) IRStates {
	return IRStates(ir_states.Canon(string(s), string(std)))
}

// KeyValueStrings 以原始字符串形式遍历所有键值对。
// key 不带引号；value 保持其文本编码形式。
func (s IRStates) KeyValueStrings(yield func(string, string) bool) {
	ir_states.KeyValueStrings((string(s)))(yield)
}

// KeyValues 以类型化值遍历所有键值对：
// string、int32，或用 int8(0/1) 表示 false/true。
func (s IRStates) KeyValues(yield func(string, any) bool) {
	ir_states.KeyValues((string(s)))(yield)
}

// Get 在 states 中查找指定 key，并返回其类型化值。
func (s IRStates) Get(key string) (value any, found bool) {
	return ir_states.Get(string(s), key)
}

// Set 插入或更新一个键值对，并返回规范化后的 states。
func (s IRStates) Set(key string, value any) IRStates {
	return IRStates(ir_states.Set(string(s), key, value))
}

// Delete 删除指定 key，并返回规范化后的 states。
func (s IRStates) Delete(key string) IRStates {
	return IRStates(ir_states.Delete(string(s), key))
}

// CastToString 将所有 value 都转换为带引号的字符串值。
// true/false 会转为 "1"/"0"，整数会转为十进制字符串。
func (s IRStates) CastToString() IRStates {
	return IRStates(ir_states.CastToString(string(s)))
}

// CompareResult 表示两个 IRStates 的模糊比较结果。
type CompareResult struct {
	// ExactlySame 仅在原始 states 字节级完全相同时为 true。
	ExactlySame bool
	// Same 表示双方都存在且映射为字符串后值相同的键数量。
	Same int
	// Different 表示双方都存在但值不同的键数量。
	Different int
	// Missing 表示 s1 缺失而 s2 存在的键数量。
	Missing int
	// Extra 表示 s1 存在而 s2 缺失的键数量。
	Extra int
}

// FuzzyCompare 在 CastToString 映射后，按键值映射比较两个 states。
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
// DecodeToHL 将 states 解码到高层 NBT 编码器。
func (s IRStates) DecodeToHL(enc nbt.EncoderHighLevel) error {
	return decoder_bedrock_block_props.Decode(s, enc, true)
}

// DecodeToLL 将 states 解码到低层 NBT 编码器。
func (s IRStates) DecodeToLL(enc nbt.Encoder) error {
	return decoder_bedrock_block_props.Decode(s, decoder_shim.NewDecoderShim(enc), true)
}

// DecodeToHL 将 block name 和 states 解码到高层 NBT 编码器。
func (b IRBlock) DecodeToHL(encoder nbt.EncoderHighLevel) (err error) {
	return decodeToHL(b, encoder)
}

// DecodeToLL 将 block name 和 states 解码到低层 NBT 编码器。
func (b IRBlock) DecodeToLL(encoder nbt.Encoder) (err error) {
	return decodeToHL(b, decoder_shim.NewDecoderShim(encoder))
}

// encoders
// StateEncoder 用于从解码事件增量构建 IRStates。
type StateEncoder struct {
	encoder_bedrock_block_props.BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]
}

// States 返回编码后的 states，并保证为 tidy 规范形式。
func (enc *StateEncoder) States() IRStates {
	return IRStates(enc.BedrockBlockPropsEncoder.W.Bytes()).Tidy()
}

// NewIRStateEncoder 创建新的 StateEncoder。
func NewIRStateEncoder() *StateEncoder {
	return &StateEncoder{}
}

// BlockEncoder 用于从解码事件增量构建 IRBlock。
type BlockEncoder struct {
	aux      uint8
	propsEnc encoder_bedrock_block_props.BedrockBlockPropsEncoder[*seq_io.SliceGrowWriter]
	block    string
}

// NewIRBlockEncoder 创建新的 BlockEncoder。
func NewIRBlockEncoder() *BlockEncoder {
	return &BlockEncoder{}
}

// Block 返回编码后的方块文本。
func (e *BlockEncoder) Block() IRBlock {
	return IRBlock(e.block)
}

// NewIRBlockFromGo 根据方块名和 Go map 属性构建 IRBlock。
// 当 props 为 nil 时，会按空 states map 处理。
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
