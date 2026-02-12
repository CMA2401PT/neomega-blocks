package encoder_block_hash

const (
	FNV1_32_INIT  uint32 = 0x811C9DC5
	FNV1_PRIME_32 uint32 = 0x01000193
)

type Fnv1a32Writer struct {
	Hash uint32
}

func NewFnv1a32Writer() *Fnv1a32Writer {
	return &Fnv1a32Writer{
		Hash: FNV1_32_INIT,
	}
}
func (w *Fnv1a32Writer) Write(bs []byte) (err error) {
	for _, b := range bs {
		w.Hash ^= uint32(b)
		w.Hash *= FNV1_PRIME_32
	}
	return nil
}
func (w *Fnv1a32Writer) WriteByte(b byte) error {
	w.Hash ^= uint32(b)
	w.Hash *= FNV1_PRIME_32
	return nil
}
