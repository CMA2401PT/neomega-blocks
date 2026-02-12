package num_lookup

import (
	"strings"
	"testing"
)

func TestBuildStaticFromPairs(t *testing.T) {
	pairs := []Pair[uint16, uint32]{
		{K: 7, V: 70},
		{K: 3, V: 30},
		{K: 7, V: 71},
		{K: 1, V: 10},
	}

	m := BuildStaticFromPairs(pairs, true)
	if m.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", m.Len())
	}

	if got, ok := m.Get(1); !ok || got != 10 {
		t.Fatalf("Get(1) = (%d, %v), want (10, true)", got, ok)
	}
	if got, ok := m.Get(3); !ok || got != 30 {
		t.Fatalf("Get(3) = (%d, %v), want (30, true)", got, ok)
	}
	if got, ok := m.Get(7); !ok || got != 71 {
		t.Fatalf("Get(7) = (%d, %v), want (71, true)", got, ok)
	}
}

func TestBuildStaticFromPairsDuplicatePanics(t *testing.T) {
	defer func() {
		p := recover()
		if p == nil {
			t.Fatal("expected panic, got nil")
		}
		if msg, ok := p.(string); !ok || msg != "duplicate key" {
			t.Fatalf("unexpected panic: %v", p)
		}
	}()
	_ = BuildStaticFromPairs([]Pair[uint16, uint16]{
		{K: 1, V: 10},
		{K: 1, V: 11},
	}, false)
}

func TestStaticNumMapDumpLoad(t *testing.T) {
	m := BuildStaticFromPairs([]Pair[uint16, uint32]{
		{K: 100, V: 1000},
		{K: 2, V: 20},
		{K: 99, V: 990},
	}, false)

	data := m.Dump()
	loaded := LoadStaticNumMap[uint16, uint32](data)

	if loaded.Len() != 3 {
		t.Fatalf("loaded.Len() = %d, want 3", loaded.Len())
	}
	if got, ok := loaded.Get(2); !ok || got != 20 {
		t.Fatalf("loaded.Get(2) = (%d, %v), want (20, true)", got, ok)
	}
	if got, ok := loaded.Get(99); !ok || got != 990 {
		t.Fatalf("loaded.Get(99) = (%d, %v), want (990, true)", got, ok)
	}
	if got, ok := loaded.Get(100); !ok || got != 1000 {
		t.Fatalf("loaded.Get(100) = (%d, %v), want (1000, true)", got, ok)
	}
	if _, ok := loaded.Get(101); ok {
		t.Fatalf("loaded.Get(101) found = true, want false")
	}
}

func TestLoadStaticNumMapTypeMismatchPanics(t *testing.T) {
	base := BuildStaticFromPairs([]Pair[uint16, uint16]{
		{K: 1, V: 10},
	}, false)
	data := base.Dump()

	defer func() {
		p := recover()
		if p == nil {
			t.Fatal("expected panic, got nil")
		}
		if !strings.Contains(p.(error).Error(), "type size mismatch") {
			t.Fatalf("unexpected panic: %v", p)
		}
	}()
	_ = LoadStaticNumMap[uint16, uint32](data)
}

func TestTwoLayerNumMapSetGetFrozenDumpLoad(t *testing.T) {
	var m TwoLayerNumMap[uint16, uint16]

	m.Set(1, 10)
	m.Set(2, 20)
	m.Frozen()

	if len(m.dynamic) != 0 {
		t.Fatalf("dynamic should be empty after Frozen, got len=%d", len(m.dynamic))
	}
	if got, ok := m.Get(1); !ok || got != 10 {
		t.Fatalf("Get(1) = (%d, %v), want (10, true)", got, ok)
	}
	if got, ok := m.Get(2); !ok || got != 20 {
		t.Fatalf("Get(2) = (%d, %v), want (20, true)", got, ok)
	}

	m.Set(2, 20)
	if len(m.dynamic) != 0 {
		t.Fatalf("setting same value should not create dynamic override, len=%d", len(m.dynamic))
	}

	m.Set(2, 99)
	m.Set(3, 30)
	if got, ok := m.Get(2); !ok || got != 99 {
		t.Fatalf("Get(2 override) = (%d, %v), want (99, true)", got, ok)
	}
	if got, ok := m.Get(3); !ok || got != 30 {
		t.Fatalf("Get(3) = (%d, %v), want (30, true)", got, ok)
	}

	dumped := m.Dump()
	if len(m.dynamic) == 0 {
		t.Fatalf("Dump uses value receiver and should not mutate original map")
	}

	loaded := LoadTwoLayerNumMap[uint16, uint16](dumped)
	if got, ok := loaded.Get(1); !ok || got != 10 {
		t.Fatalf("loaded.Get(1) = (%d, %v), want (10, true)", got, ok)
	}
	if got, ok := loaded.Get(2); !ok || got != 99 {
		t.Fatalf("loaded.Get(2) = (%d, %v), want (99, true)", got, ok)
	}
	if got, ok := loaded.Get(3); !ok || got != 30 {
		t.Fatalf("loaded.Get(3) = (%d, %v), want (30, true)", got, ok)
	}
	if len(loaded.dynamic) != 0 {
		t.Fatalf("loaded dynamic should be empty, got len=%d", len(loaded.dynamic))
	}
}
