package encoder_block_hash

import (
	encoder_nbt "neomega-nbt/encoder-nbt"
)

type Encoder struct {
	encoder_nbt.NBTEncoder[
		*Fnv1a32Writer,
		encoder_nbt.LittleEndian[*Fnv1a32Writer],
	]
	w *Fnv1a32Writer
}

func NewEncoder() Encoder {
	w := NewFnv1a32Writer()
	enc := encoder_nbt.NewEncoder[
		*Fnv1a32Writer,
		encoder_nbt.LittleEndian[*Fnv1a32Writer],
	](w)
	return Encoder{
		NBTEncoder: enc,
		w:          w,
	}
}

func (enc Encoder) Hash() uint32 {
	return enc.w.Hash
}

func (enc Encoder) Reset() {
	enc.w.Hash = FNV1_32_INIT
}
