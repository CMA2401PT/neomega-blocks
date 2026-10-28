package num_lookup

import (
	"iter"
	"maps"
)

type TwoLayerNumMap[K, V ~uint8 | ~uint16 | ~uint32] struct {
	static  StaticNumMap[K, V]
	dynamic map[K]V
}

func NewTwoLayerNumMap[K, V ~uint8 | ~uint16 | ~uint32]() *TwoLayerNumMap[K, V] {
	return &TwoLayerNumMap[K, V]{
		static:  StaticNumMap[K, V]{keys: make([]K, 0), vals: make([]V, 0)},
		dynamic: map[K]V{},
	}
}

func (m TwoLayerNumMap[K, V]) Get(k K) (v V, ok bool) {
	// 先尝试从 dynamic中找，找不到就去static 中找
	if m.dynamic != nil {
		if v, ok := m.dynamic[k]; ok {
			return v, true
		}
	}
	return m.static.Get(k)
}

func (m *TwoLayerNumMap[K, V]) Set(k K, v V) {
	// 先看看现在有没有这个值，有且相同就不管了
	// 有但不同的话只修改 dynamic，不要动 static
	// 没有的话就增加
	if cur, ok := m.Get(k); ok && cur == v {
		return
	}
	if m.dynamic == nil {
		m.dynamic = make(map[K]V)
	}
	m.dynamic[k] = v
}

func (m *TwoLayerNumMap[K, V]) Frozen() {
	// 重构并更新 static，然后丢掉 dynamic
	if len(m.dynamic) == 0 {
		return
	}

	pairs := make([]Pair[K, V], 0, m.static.Len()+len(m.dynamic))
	for i := range m.static.keys {
		k := m.static.keys[i]
		if v, ok := m.dynamic[k]; ok {
			pairs = append(pairs, Pair[K, V]{K: k, V: v})
			delete(m.dynamic, k)
			continue
		}
		pairs = append(pairs, Pair[K, V]{K: k, V: m.static.vals[i]})
	}
	for k, v := range m.dynamic {
		pairs = append(pairs, Pair[K, V]{K: k, V: v})
	}

	m.static = BuildStaticFromPairs(pairs, false)
	m.dynamic = nil
}

func (m TwoLayerNumMap[K, V]) Dump() []uint8 {
	// Frozen removes merged keys; the value receiver still shares the original map.
	m.dynamic = maps.Clone(m.dynamic)
	m.Frozen()
	return m.static.Dump()
}

func LoadTwoLayerNumMap[K, V ~uint8 | ~uint16 | ~uint32](data []uint8) TwoLayerNumMap[K, V] {
	return TwoLayerNumMap[K, V]{
		static:  LoadStaticNumMap[K, V](data),
		dynamic: nil,
	}
}

// All iterates the merged view without freezing or mutating either layer.
// Iteration order is unspecified; callers must not modify the map while reading.
func (m TwoLayerNumMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i, key := range m.static.keys {
			value := m.static.vals[i]
			if override, ok := m.dynamic[key]; ok {
				value = override
			}
			if !yield(key, value) {
				return
			}
		}
		for key, value := range m.dynamic {
			if _, exists := m.static.Get(key); !exists {
				if !yield(key, value) {
					return
				}
			}
		}
	}
}
