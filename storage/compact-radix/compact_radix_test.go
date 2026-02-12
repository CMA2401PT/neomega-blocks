package compact_radix

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompactRadixSetGetAndOverwrite(t *testing.T) {
	r := NewCommonCompactRadix[uint16, uint16]()

	idxA := r.Set([]byte("cat"), 10)
	idxB := r.Set([]byte("car"), 20)
	idxC := r.Set([]byte("dog"), 30)

	if idxA == idxB || idxA == idxC || idxB == idxC {
		t.Fatalf("expected unique node indexes, got: %d %d %d", idxA, idxB, idxC)
	}

	cases := []struct {
		key   string
		value uint16
		found bool
	}{
		{key: "cat", value: 10, found: true},
		{key: "car", value: 20, found: true},
		{key: "dog", value: 30, found: true},
		{key: "cab", value: 0, found: false},
		{key: "cow", value: 0, found: false},
	}

	for _, tc := range cases {
		got, ok := r.Get([]byte(tc.key))
		if ok != tc.found {
			t.Fatalf("Get(%q) found = %v, want %v", tc.key, ok, tc.found)
		}
		if ok && got != tc.value {
			t.Fatalf("Get(%q) value = %d, want %d", tc.key, got, tc.value)
		}
	}

	idxA2 := r.Set([]byte("cat"), 99)
	if idxA2 != idxA {
		t.Fatalf("overwrite should reuse same node index: got %d, want %d", idxA2, idxA)
	}
	got, ok := r.Get([]byte("cat"))
	if !ok || got != 99 {
		t.Fatalf("Get(%q) = (%d, %v), want (99, true)", "cat", got, ok)
	}
}

func TestCompactRadixSetPanicsForPrefixConflicts(t *testing.T) {
	t.Run("new key is prefix of existing key", func(t *testing.T) {
		r := NewCommonCompactRadix[uint16, uint16]()
		r.Set([]byte("abc"), 1)

		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("expected panic, got nil")
			}
			if !strings.Contains(p.(error).Error(), "is a prefix of other key") {
				t.Fatalf("unexpected panic: %v", p)
			}
		}()
		r.Set([]byte("ab"), 2)
	})

	t.Run("existing key is prefix of new key", func(t *testing.T) {
		r := NewCommonCompactRadix[uint16, uint16]()
		r.Set([]byte("ab"), 1)

		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("expected panic, got nil")
			}
			if !strings.Contains(p.(error).Error(), "is a prefix of other key") {
				t.Fatalf("unexpected panic: %v", p)
			}
		}()
		r.Set([]byte("abc"), 2)
	})

	t.Run("empty key always panics", func(t *testing.T) {
		r := NewCommonCompactRadix[uint16, uint16]()
		defer func() {
			p := recover()
			if p == nil {
				t.Fatal("expected panic, got nil")
			}
			if got := p.(string); got != "empty key is always a prefix" {
				t.Fatalf("unexpected panic message: %q", got)
			}
		}()
		r.Set([]byte(""), 1)
	})
}

func TestCompactRadixRebuild(t *testing.T) {
	r := NewCommonCompactRadix[uint16, uint16]()

	keys := []string{"cat", "car", "dog", "dot", "zebra"}
	indexByKey := map[string]uint16{}
	for i, k := range keys {
		indexByKey[k] = r.Set([]byte(k), uint16(i+1))
	}

	for _, k := range keys {
		rebuilt := r.Rebuild(indexByKey[k])
		if string(rebuilt) != k {
			t.Fatalf("Rebuild(%d) = %q, want %q", indexByKey[k], string(rebuilt), k)
		}
	}
}

func TestCompactRadixIter(t *testing.T) {
	r := NewCommonCompactRadix[uint16, uint16]()
	want := map[string]uint16{
		"car":   1,
		"cat":   2,
		"dog":   3,
		"zebra": 4,
	}
	for k, v := range want {
		r.Set([]byte(k), v)
	}

	got := map[string]uint16{}
	r.Iter(func(k []uint8, v uint16) bool {
		got[string(k)] = v
		return true
	})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Iter output mismatch: got %#v, want %#v", got, want)
	}

	count := 0
	r.Iter(func(_ []uint8, _ uint16) bool {
		count++
		return false
	})
	if count != 1 {
		t.Fatalf("Iter should stop when callback returns false, got count=%d", count)
	}
}
