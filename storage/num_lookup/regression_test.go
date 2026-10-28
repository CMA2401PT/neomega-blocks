package num_lookup

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestLoadUnalignedStaticMap(t *testing.T) {
	original := BuildStaticFromPairs([]Pair[uint16, uint32]{{K: 7, V: 700}, {K: 1, V: 100}}, false)
	data := original.Dump()
	// A byte slice (including go:embed data) need not start at a uint32-aligned address.
	backing := make([]uint32, (len(data)+4)/4)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&backing[0])), len(backing)*4)
	unaligned := raw[1 : 1+len(data)]
	copy(unaligned, data)
	loaded := LoadStaticNumMap[uint16, uint32](unaligned)
	for _, key := range []uint16{1, 7} {
		want, _ := original.Get(key)
		if got, ok := loaded.Get(key); !ok || got != want {
			t.Fatalf("key %d: got %d,%v; want %d", key, got, ok, want)
		}
	}
}

func TestDumpPreservesPendingOverrides(t *testing.T) {
	m := NewTwoLayerNumMap[uint32, uint32]()
	m.Set(1, 10)
	m.Frozen()
	m.Set(1, 20)
	m.Set(2, 30)
	first := m.Dump()
	if got, _ := m.Get(1); got != 20 {
		t.Fatalf("Dump lost pending override: %d", got)
	}
	if second := m.Dump(); !bytes.Equal(first, second) {
		t.Fatal("repeated Dump changed contents")
	}
	m.Frozen()
	if got, _ := m.Get(1); got != 20 {
		t.Fatalf("Frozen lost override after Dump: %d", got)
	}
}

func TestOverwriteKeepsLastInputPair(t *testing.T) {
	pairs := make([]Pair[uint32, uint32], 100)
	want := map[uint32]uint32{}
	for i := range pairs {
		pairs[i] = Pair[uint32, uint32]{K: uint32(i % 7), V: uint32(i)}
		want[pairs[i].K] = pairs[i].V
	}
	m := BuildStaticFromPairs(pairs, true)
	for key, value := range want {
		if got, _ := m.Get(key); got != value {
			t.Errorf("key %d: got %d, want last input %d", key, got, value)
		}
	}
}

func TestAllReadsTheMergedViewWithoutMutatingIt(t *testing.T) {
	original := NewTwoLayerNumMap[uint32, uint32]()
	original.Set(1, 10)
	original.Set(2, 20)
	original.Frozen()
	original.Set(2, 25)
	original.Set(3, 30)
	got := map[uint32]uint32{}
	for key, value := range original.All() {
		if _, exists := got[key]; exists {
			t.Fatal("duplicate key")
		}
		got[key] = value
	}
	if len(got) != 3 || got[1] != 10 || got[2] != 25 || got[3] != 30 {
		t.Fatalf("merged entries: %v", got)
	}
	if value, _ := original.Get(2); value != 25 {
		t.Fatal("iteration changed the map")
	}
}
