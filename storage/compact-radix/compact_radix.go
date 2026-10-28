package compact_radix

import (
	"fmt"
	"strings"
	"unsafe"
)

// 读取和查找的效率都非常低下，但是是最节约空间的
// 每个key 都不允许是其他 key 的前缀，也不允许有删除
// 这样保证每个set产生的index一定是叶子节点，且一定有值
type CompactRadix[EP, NP ~uint8 | ~uint16 | ~uint32, Lookup EdgeLookup[EP]] struct {
	Lookup Lookup
	nodes  []node[EP, NP]
	// debugMap map[NP][]uint8
}

type CommonCompactRadix[EP, NP ~uint8 | ~uint16 | ~uint32] struct {
	CompactRadix[EP, NP, *ByteEdgeLookup[EP]]
}

func NewCommonCompactRadix[EP, NP ~uint8 | ~uint16 | ~uint32]() *CommonCompactRadix[EP, NP] {
	r := &CommonCompactRadix[EP, NP]{}
	r.Lookup = NewByteEdgeLookup[EP]()
	return r
}

func NewCompactRadixWithLookup[EP, NP ~uint8 | ~uint16 | ~uint32, Lookup EdgeLookup[EP]](lookup Lookup) *CompactRadix[EP, NP, Lookup] {
	r := &CompactRadix[EP, NP, Lookup]{}
	r.Lookup = lookup
	return r
}

type node[EP, NP ~uint8 | ~uint16 | ~uint32] struct {
	start, end                 EP
	parentChildXOR, siblingXOR NP
	parentOrValue              NP
}

func (r *CompactRadix[EP, NP, Lookup]) DumpNodes() []uint8 {
	if len(r.nodes) == 0 {
		return nil
	}
	nodeSize := unsafe.Sizeof(node[EP, NP]{})
	out := unsafe.Slice(
		(*uint8)(unsafe.Pointer(unsafe.SliceData(r.nodes))),
		len(r.nodes)*int(nodeSize),
	)
	return out
}

func (r *CompactRadix[EP, NP, Lookup]) LoadNodes(data []uint8) {
	if len(data) == 0 {
		r.nodes = nil
		return
	}
	nodeSize := unsafe.Sizeof(node[EP, NP]{})
	if len(data)%int(nodeSize) != 0 {
		panic(fmt.Errorf("invalid data size %d for node size %d", len(data), nodeSize))
	}
	align := unsafe.Alignof(node[EP, NP]{})
	if uintptr(unsafe.Pointer(unsafe.SliceData(data)))%align != 0 {
		// Serialized offsets do not guarantee alignment of the backing byte slice.
		r.nodes = make([]node[EP, NP], len(data)/int(nodeSize))
		copy(r.DumpNodes(), data)
		return
	}
	r.nodes = unsafe.Slice(
		(*node[EP, NP])(unsafe.Pointer(unsafe.SliceData(data))),
		len(data)/int(nodeSize),
	)
}

func (r *CompactRadix[EP, NP, Lookup]) edgeGetIndex(path []uint8) (start, end EP) {
	return r.Lookup.GetStableIndex(path)
}

func (r *CompactRadix[EP, NP, Lookup]) appendNode(n node[EP, NP]) (nodeIdx NP) {
	if r.nodes == nil {
		r.nodes = make([]node[EP, NP], 0)
	}
	idx := len(r.nodes)
	if idx+1 == int(^(NP(0))) {
		panic("nodes size overflow")
	}
	r.nodes = append(r.nodes, n)
	return NP(idx)
}

func (r *CompactRadix[EP, NP, Lookup]) edgeGetSlice(start, end EP) []uint8 {
	return r.Lookup.GetPath(start, end)
}

func (r *CompactRadix[EP, NP, Lookup]) Set(key []uint8, value NP) (nodeIdx NP) {
	origKey := key
	if len(r.nodes) == 0 {
		n := node[EP, NP]{
			start: 0, end: 0,
			parentChildXOR: 0,
			parentOrValue:  ^NP(0),
		}
		r.appendNode(n)
		// r.debugMap = map[NP][]uint8{}
	}
	if len(key) == 0 {
		// r.nodes[0].hasValue = true
		// r.nodes[0].value = value
		panic("empty key is always a prefix")
	}
	// 开始时，所有节点都是 空 的子节点
	parentIdx := NP(0)                                   // 上一级节点的位置
	firstSibling := r.nodes[0].parentChildXOR ^ (^NP(0)) // 这一级第一个节点位置
	for {
		if firstSibling == ^NP(0) {
			// 如果这层什么节点都没有，则创建第一个节点
			s, e := r.edgeGetIndex(key)
			n := node[EP, NP]{
				parentOrValue: (value),
				// hasValue:  true,
				// firstByte: key[0],
				start: s, end: e,
				parentChildXOR: parentIdx ^ (^NP(0)),
				siblingXOR:     0,
			}
			thisNode := r.appendNode(n)
			parentNode := r.nodes[parentIdx]
			parentNode.parentChildXOR = parentNode.parentChildXOR ^ (^NP(0)) ^ thisNode
			r.nodes[parentIdx] = parentNode
			// r.debugMap[thisNode] = origKey
			return thisNode
		}
		brotherNodeIdx, nodeIdx, label, ok := r.findEdge(firstSibling, key[0]) // 从这层的第一个节点开始找起
		if !ok {
			// 应该在最后一个兄弟节点后挂上它
			s, e := r.edgeGetIndex(key)
			n := node[EP, NP]{
				parentOrValue: (value),
				// hasValue:  true,
				// firstByte: key[0],
				start: s, end: e,
				parentChildXOR: parentIdx ^ (^NP(0)),
				siblingXOR:     nodeIdx ^ (^NP(0)),
			}
			thisIdx := r.appendNode(n)
			brotherNode := r.nodes[nodeIdx]
			brotherNode.siblingXOR = (brotherNode.siblingXOR ^ (^NP(0))) ^ thisIdx //令兄节点指向它
			r.nodes[nodeIdx] = brotherNode
			// _, ok = r.debugMap[thisIdx]
			// if ok {
			// 	panic("should be an exist node")
			// }
			// r.debugMap[thisIdx] = origKey
			return thisIdx
		}
		n := r.nodes[nodeIdx]
		common := commonPrefixLen(label, key)
		switch {
		case common == len(label) && common == len(key):
			// 完全匹配了这个节点
			firstSibling = n.parentChildXOR ^ parentIdx
			if firstSibling != ^NP(0) {
				panic(fmt.Errorf("key %v(%v) is a prefix of other key", string(origKey), origKey))
			}
			// n.hasValue = true
			n.parentOrValue = value
			r.nodes[nodeIdx] = n
			// _, ok = r.debugMap[nodeIdx]
			// if !ok {
			// 	r.debugMap[nodeIdx] = origKey
			// } else if !bytes.Equal(r.debugMap[nodeIdx], origKey) {
			// 	panic("matched node index contains different debug key")
			// }
			return nodeIdx
		case common == len(label):
			// 前缀匹配了，需要检查这个节点的子节点
			firstSibling = n.parentChildXOR ^ parentIdx
			parentIdx = nodeIdx
			key = key[common:]
			if firstSibling == ^NP(0) {
				panic(fmt.Errorf("key %v(%v) is a prefix of other key", string(origKey), origKey))
			}
			continue
		default:
			// 需要分裂节点
			prefixStart, prefixEnd := r.edgeGetIndex(label[:common])
			suffixStart, suffixEnd := r.edgeGetIndex(label[common:])
			// 这个节点将成为现有节点和即将插入节点的新的共同前缀节点
			newPrefixNode := node[EP, NP]{
				// hasValue:  false,
				// firstByte: key[0],
				start: prefixStart, end: prefixEnd,
				parentChildXOR: parentIdx ^ nodeIdx,
				siblingXOR:     n.siblingXOR,
				parentOrValue:  parentIdx,
			}

			// 修改其亲节点指向
			prefixNodeIdx := r.appendNode(newPrefixNode)
			if firstSibling == nodeIdx {
				parentNode := r.nodes[parentIdx]
				parentNode.parentChildXOR = parentNode.parentChildXOR ^ nodeIdx ^ prefixNodeIdx
				r.nodes[parentIdx] = parentNode
			}
			// 修改其兄弟节点指向
			if brotherNodeIdx != ^(NP(0)) {
				brotherNode := r.nodes[brotherNodeIdx]
				brotherNode.siblingXOR = brotherNode.siblingXOR ^ nodeIdx ^ prefixNodeIdx
				r.nodes[brotherNodeIdx] = brotherNode
			}
			brotherNodeIdx = n.siblingXOR ^ brotherNodeIdx
			if brotherNodeIdx != ^(NP(0)) {
				brotherNode := r.nodes[brotherNodeIdx]
				brotherNode.siblingXOR = brotherNode.siblingXOR ^ nodeIdx ^ prefixNodeIdx
				r.nodes[brotherNodeIdx] = brotherNode
			}

			// n.firstByte = label[common]
			oldFirstChild := n.parentChildXOR ^ parentIdx
			n.start = suffixStart
			n.end = suffixEnd
			n.parentChildXOR = (n.parentChildXOR ^ parentIdx) ^ prefixNodeIdx
			// If n is an internal node, parentOrValue stores parent index and must be rewired.
			if oldFirstChild != ^NP(0) {
				n.parentOrValue = prefixNodeIdx
			}
			n.siblingXOR = 0
			r.nodes[nodeIdx] = n
			parentIdx = prefixNodeIdx
			firstSibling = nodeIdx
			key = key[common:]
			if len(key) == 0 {
				panic(fmt.Errorf("key %v(%v) is a prefix of other key", string(origKey), origKey))
				// newPrefixNode.hasValue = true
				// newPrefixNode.value = value
				// r.nodes[prefixNodeIdx] = newPrefixNode
				// return prefixNodeIdx
			}
		}
	}
}

func commonPrefixLen(a, b []byte) int {
	max := len(a)
	if len(b) < max {
		max = len(b)
	}
	for i := 0; i < max; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}

func (r *CompactRadix[EP, NP, Lookup]) findEdge(nodeIdx NP, firstByte uint8) (NP, NP, []uint8, bool) {
	brotherNode := ^NP(0) // 这个节点（nodeIdx）的兄节点
	for {
		n := r.nodes[nodeIdx]
		slice := r.edgeGetSlice(n.start, n.end)
		if slice[0] == firstByte {
			return brotherNode, nodeIdx, slice, true
		}
		nextNodeIdx := n.siblingXOR ^ brotherNode // 下一个要检查的节点
		if nextNodeIdx == ^NP(0) {                // 没有下一个节点
			return brotherNode, nodeIdx, nil, false
		}
		brotherNode = nodeIdx
		nodeIdx = nextNodeIdx
	}
}

func (r *CompactRadix[EP, NP, Lookup]) Get(key []uint8) (value NP, found bool) {
	parentIdx := NP(0)                                   // 上一级节点的位置
	firstSibling := r.nodes[0].parentChildXOR ^ (^NP(0)) // 这一级第一个节点位置
	for {
		if firstSibling == ^NP(0) {
			return value, false
		}
		_, nodeIdx, label, ok := r.findEdge(firstSibling, key[0]) // 从这层的第一个节点开始找起
		if !ok {
			return value, false
		}
		n := r.nodes[nodeIdx]
		common := commonPrefixLen(label, key)
		switch {
		case common == len(label) && common == len(key):
			// 完全匹配了这个节点
			return n.parentOrValue, true
		case common == len(label):
			// 前缀匹配了，需要检查这个节点的子节点
			firstSibling = n.parentChildXOR ^ parentIdx
			parentIdx = nodeIdx
			key = key[common:]
			continue
		default:
			return value, false
		}
	}
}

func (r *CompactRadix[EP, NP, Lookup]) SetByIdx(nodeIdx NP, value NP) {
	r.nodes[nodeIdx].parentOrValue = value
}

func (r *CompactRadix[EP, NP, Lookup]) GetByIdx(nodeIdx NP) NP {
	return r.nodes[nodeIdx].parentOrValue
}
func (r *CompactRadix[EP, NP, Lookup]) Rebuild(nodeIdx NP) []uint8 {
	// 从 nodeIdx 重新构建整个key(path)
	out := make([]uint8, 0)
	n := r.nodes[nodeIdx]
	parentIdx := n.parentChildXOR ^ (^NP(0))
	for {
		slice := r.edgeGetSlice(n.start, n.end)
		for i := range slice {
			out = append(out, slice[len(slice)-i-1])
		}
		if parentIdx == ^NP(0) {
			break
		}
		n = r.nodes[parentIdx]
		parentIdx = n.parentOrValue
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	// if expect, ok := r.debugMap[nodeIdx]; ok && !bytes.Equal(expect, out) {
	// 	panic("rebuild error")
	// }
	return out
}

func (r CompactRadix[EP, NP, Lookup]) Iter(yield func([]uint8, NP) bool) {
	var writeChildrens func(path []uint8, parentIdx, nodeIdx NP) bool
	writeChildrens = func(path []uint8, parentIdx, nodeIdx NP) bool {
		brotherIdx := ^NP(0)
		for {
			n := r.nodes[nodeIdx]
			frag := r.edgeGetSlice(n.start, n.end)
			fullPath := append(path, frag...)
			firstChild := n.parentChildXOR ^ parentIdx
			if firstChild != ^NP(0) {
				if !writeChildrens(fullPath, nodeIdx, firstChild) {
					return false
				}
			} else {
				if !yield(fullPath, n.parentOrValue) {
					return false
				}
			}
			nextIdx := n.siblingXOR ^ brotherIdx
			if nextIdx == ^NP(0) {
				break
			}
			brotherIdx = nodeIdx
			nodeIdx = nextIdx
		}
		return true
	}
	writeChildrens([]uint8{}, ^NP(0), 0)
}

func (r CompactRadix[EP, NP, Lookup]) String() string {
	s := strings.Builder{}
	for fullPath, value := range r.Iter {
		s.WriteString(fmt.Sprintf("%v:%v\n", string(fullPath), value))
	}
	return s.String()
}
