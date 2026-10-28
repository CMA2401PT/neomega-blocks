package num_lookup

import (
	"encoding/binary"
	"fmt"
	"sort"
	"unsafe"
)

type Pair[K ~uint8 | ~uint16 | ~uint32, V any] struct {
	K K
	V V
}

type StaticNumMap[K ~uint8 | ~uint16 | ~uint32, V any] struct {
	keys []K
	vals []V
}

// BuildStaticFromPairs builds a compact immutable map from pairs.
// If allowOverwrite is true, later duplicates overwrite earlier ones (after sorting by key, last wins).
func BuildStaticFromPairs[K ~uint8 | ~uint16 | ~uint32, V any](pairs []Pair[K, V], allowOverwrite bool) StaticNumMap[K, V] {
	if len(pairs) == 0 {
		return StaticNumMap[K, V]{}
	}

	// Preserve input order for equal keys so the last supplied value wins.
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].K < pairs[j].K })

	keys := make([]K, 0, len(pairs))
	vals := make([]V, 0, len(pairs))

	var last K
	haveLast := false
	for _, p := range pairs {
		if haveLast && p.K == last {
			if !allowOverwrite {
				panic("duplicate key")
			}
			// overwrite last value
			vals[len(vals)-1] = p.V
			continue
		}
		haveLast = true
		last = p.K
		keys = append(keys, p.K)
		vals = append(vals, p.V)
	}

	return StaticNumMap[K, V]{keys: keys, vals: vals}
}

func (m StaticNumMap[K, V]) Len() int { return len(m.keys) }

func (m StaticNumMap[K, V]) Get(k K) (v V, ok bool) {
	i := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= k })
	if i < len(m.keys) && m.keys[i] == k {
		return m.vals[i], true
	}
	var zero V
	return zero, false
}

func (m StaticNumMap[K, V]) Dump() []uint8 {
	const headerSize = 8
	const staticMagic = "SNM1"

	keySize := int(unsafe.Sizeof(*new(K)))
	valSize := int(unsafe.Sizeof(*new(V)))
	keyAlign := int(unsafe.Alignof(*new(K)))
	valAlign := int(unsafe.Alignof(*new(V)))
	count := len(m.keys)

	if count != len(m.vals) {
		panic("invalid static map: keys/vals length mismatch")
	}

	keysBytesLen := count * keySize
	valsBytesLen := count * valSize

	total := headerSize + 4
	total += (keyAlign - (total % keyAlign)) % keyAlign
	keysOffset := total
	total += keysBytesLen
	total += (valAlign - (total % valAlign)) % valAlign
	valsOffset := total
	total += valsBytesLen

	out := make([]byte, total)
	copy(out[:4], []byte(staticMagic))
	binary.LittleEndian.PutUint32(out[4:8], uint32(count))
	out[8] = byte(keySize)
	out[9] = byte(valSize)
	out[10] = byte(keyAlign)
	out[11] = byte(valAlign)

	if count > 0 {
		keyBytes := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(m.keys))), keysBytesLen)
		valBytes := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(m.vals))), valsBytesLen)
		copy(out[keysOffset:keysOffset+keysBytesLen], keyBytes)
		copy(out[valsOffset:valsOffset+valsBytesLen], valBytes)
	}

	return out
}

func LoadStaticNumMap[K ~uint8 | ~uint16 | ~uint32, V any](data []uint8) StaticNumMap[K, V] {
	const headerSize = 8
	const staticMagic = "SNM1"

	if len(data) < 12 {
		panic("invalid static map dump: too short")
	}
	if string(data[:4]) != staticMagic {
		panic("invalid static map dump: bad magic")
	}

	count := int(binary.LittleEndian.Uint32(data[4:8]))
	keySize := int(data[8])
	valSize := int(data[9])
	keyAlign := int(data[10])
	valAlign := int(data[11])

	expectedKeySize := int(unsafe.Sizeof(*new(K)))
	expectedValSize := int(unsafe.Sizeof(*new(V)))
	expectedKeyAlign := int(unsafe.Alignof(*new(K)))
	expectedValAlign := int(unsafe.Alignof(*new(V)))

	if keySize != expectedKeySize || valSize != expectedValSize {
		panic(fmt.Errorf("invalid static map dump: type size mismatch: key=%d/%d val=%d/%d", keySize, expectedKeySize, valSize, expectedValSize))
	}
	if keyAlign != expectedKeyAlign || valAlign != expectedValAlign {
		panic(fmt.Errorf("invalid static map dump: type alignment mismatch: key=%d/%d val=%d/%d", keyAlign, expectedKeyAlign, valAlign, expectedValAlign))
	}
	if count < 0 {
		panic("invalid static map dump: negative count")
	}

	keysBytesLen := count * keySize
	valsBytesLen := count * valSize

	off := headerSize + 4
	off += (keyAlign - (off % keyAlign)) % keyAlign
	keysOffset := off
	off += keysBytesLen
	off += (valAlign - (off % valAlign)) % valAlign
	valsOffset := off
	off += valsBytesLen

	if off != len(data) {
		panic("invalid static map dump: bad payload size")
	}
	if count == 0 {
		return StaticNumMap[K, V]{}
	}

	return StaticNumMap[K, V]{
		keys: loadSlice[K](data[keysOffset:keysOffset+keysBytesLen], count),
		vals: loadSlice[V](data[valsOffset:valsOffset+valsBytesLen], count),
	}
}

// The dump aligns offsets, but its backing byte slice may itself be unaligned.
// Keep the zero-copy path when possible; otherwise copy into typed, aligned storage.
func loadSlice[T any](data []byte, count int) []T {
	ptr := unsafe.Pointer(unsafe.SliceData(data))
	if uintptr(ptr)%unsafe.Alignof(*new(T)) == 0 {
		return unsafe.Slice((*T)(ptr), count)
	}
	values := make([]T, count)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(values))), len(data))
	copy(raw, data)
	return values
}
