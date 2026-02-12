package compact_radix

import "bytes"

type EdgeLookup[EP ~uint8 | ~uint16 | ~uint32] interface {
	GetStableIndex(path []uint8) (start, end EP)
	GetPath(start, end EP) []uint8
}

// ByteEdgeLookup stores concatenated edge labels in a compact byte slice.
// It returns stable [start,end) indices for repeated paths.
type ByteEdgeLookup[EP ~uint8 | ~uint16 | ~uint32] struct {
	Data []byte
}

func NewByteEdgeLookup[EP ~uint8 | ~uint16 | ~uint32]() *ByteEdgeLookup[EP] {
	return &ByteEdgeLookup[EP]{Data: make([]byte, 0)}
}

func (*ByteEdgeLookup[EP]) New() EdgeLookup[EP] {
	return &ByteEdgeLookup[EP]{Data: make([]byte, 0)}
}

func (l *ByteEdgeLookup[EP]) Dump() []uint8 {
	return l.Data
}

func (l *ByteEdgeLookup[EP]) Load(data []uint8) {
	l.Data = data
}

func (l *ByteEdgeLookup[EP]) GetStableIndex(path []uint8) (start, end EP) {
	if len(path) == 0 {
		return 0, 0
	}
	if l.Data == nil {
		l.Data = make([]uint8, 0)
	}
	idx := bytes.Index(l.Data, path)
	if idx >= 0 {
		return EP(idx), EP(idx + len(path))
	}
	startInt := len(l.Data)
	l.Data = append(l.Data, path...)
	endInt := startInt + len(path)
	if endInt > int(^EP(0)) {
		panic("edge size overflow")
	}
	return EP(startInt), EP(endInt)
}

func (l *ByteEdgeLookup[EP]) GetPath(start, end EP) []uint8 {
	s := int(start)
	e := int(end)
	if s < 0 || e < s || e > len(l.Data) {
		return nil
	}
	return l.Data[s:e]
}
