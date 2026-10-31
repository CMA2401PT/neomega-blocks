package blockconverts

import (
	"encoding/binary"
	"fmt"
	"sort"
)

const step4StorageMagic = "BDS1"

type Step4Storage map[string][]uint32

func BuildStep4Storage(blocks []Block) Step4Storage {
	maxID := map[string]int32{}
	for _, b := range blocks {
		name := b.IRBlock.NameWithNameSpace()
		maxV, ok := maxID[name]
		if !ok || b.Data > maxV {
			maxID[name] = b.Data
		}
	}

	storage := Step4Storage{}
	for name, maxV := range maxID {
		if maxV < 0 {
			continue
		}
		storage[name] = make([]uint32, maxV+1)
	}
	for _, b := range blocks {
		name := b.IRBlock.NameWithNameSpace()
		if b.Data < 0 {
			continue
		}
		if _, ok := storage[name]; !ok {
			continue
		}
		storage[name][b.Data] = b.NetworkID
	}
	return storage
}

func (s Step4Storage) Dump() []byte {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	total := 4 + 4 // magic + count
	for _, k := range keys {
		arr := s[k]
		total += 4 + len(k)     // key len + key bytes
		total += 4 + len(arr)*4 // array len + array values
	}

	out := make([]byte, total)
	off := 0
	copy(out[off:off+4], []byte(step4StorageMagic))
	off += 4
	binary.LittleEndian.PutUint32(out[off:off+4], uint32(len(keys)))
	off += 4

	for _, k := range keys {
		arr := s[k]
		binary.LittleEndian.PutUint32(out[off:off+4], uint32(len(k)))
		off += 4
		copy(out[off:off+len(k)], []byte(k))
		off += len(k)

		binary.LittleEndian.PutUint32(out[off:off+4], uint32(len(arr)))
		off += 4
		for _, v := range arr {
			binary.LittleEndian.PutUint32(out[off:off+4], v)
			off += 4
		}
	}
	return out
}

func LoadStep4Storage(data []byte) (Step4Storage, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid step4 storage: too short")
	}
	if string(data[:4]) != step4StorageMagic {
		return nil, fmt.Errorf("invalid step4 storage: bad magic")
	}
	off := 4
	count := int(binary.LittleEndian.Uint32(data[off : off+4]))
	off += 4

	out := make(Step4Storage, count)
	for i := 0; i < count; i++ {
		if off+4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: missing key length")
		}
		keyLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		if keyLen < 0 || off+keyLen > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: bad key length")
		}
		key := string(data[off : off+keyLen])
		off += keyLen

		if off+4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: missing array length")
		}
		arrLen := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		if arrLen < 0 || off+arrLen*4 > len(data) {
			return nil, fmt.Errorf("invalid step4 storage: bad array length")
		}
		arr := make([]uint32, arrLen)
		for j := 0; j < arrLen; j++ {
			arr[j] = binary.LittleEndian.Uint32(data[off : off+4])
			off += 4
		}
		out[key] = arr
	}
	if off != len(data) {
		return nil, fmt.Errorf("invalid step4 storage: trailing bytes")
	}
	return out, nil
}
