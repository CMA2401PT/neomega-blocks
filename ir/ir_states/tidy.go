package ir_states

import (
	"sort"
	"strings"
)

// IRStates 是 ["key1"="string","key2"=true,"key3"=false,"key4"=123] 的形式
// type IRStates string

// IsTidy reports whether IRStates is already in canonical tidy form:
// no unnecessary whitespace and keys sorted by non-decreasing order.
func IsTidy(states string) bool {
	if len(states) < 2 || states[0] != '[' || states[len(states)-1] != ']' {
		return false
	}
	if states == "[]" {
		return true
	}

	prevKeyStart, prevKeyEnd := -1, -1
	i := 1
	last := len(states) - 1

	for {
		if i >= last || states[i] != '"' {
			return false
		}
		keyStart := i + 1
		i = keyStart
		for i < last && states[i] != '"' {
			i++
		}
		if i >= last {
			return false
		}
		keyEnd := i
		i++

		if i >= last || states[i] != '=' {
			return false
		}
		i++
		if i >= len(states) {
			return false
		}

		if states[i] == '"' {
			i++
			for i < last {
				if states[i] == '\\' {
					i += 2
					continue
				}
				if states[i] == '"' {
					break
				}
				i++
			}
			if i >= last || states[i] != '"' {
				return false
			}
			i++
		} else {
			valStart := i
			for i < last && states[i] != ',' && states[i] != ']' {
				i++
			}
			if valStart == i {
				return false
			}
			if isIRStateWS(states[valStart]) || isIRStateWS(states[i-1]) {
				return false
			}
		}

		if prevKeyStart != -1 && compareStateKey(states, prevKeyStart, prevKeyEnd, keyStart, keyEnd) > 0 {
			return false
		}
		prevKeyStart, prevKeyEnd = keyStart, keyEnd

		if i >= len(states) {
			return false
		}
		switch states[i] {
		case ',':
			i++
			continue
		case ']':
			return i == last
		default:
			return false
		}
	}
}

// Tidy 按移除不必要的空格 (但是 key 中的空格和字符串类型值中的空格不移除)
// 然后排序并返回结果
func Tidy(states string) string {
	if IsTidy(states) {
		return states
	}

	if len(states) < 2 || states[0] != '[' {
		return states
	}

	type stateItem struct {
		keyStart int
		keyEnd   int
		valStart int
		valEnd   int
	}

	items := make([]stateItem, 0, 8)
	i := 1
	prevKeyStart, prevKeyEnd := -1, -1
	sorted := true
	needsTidy := false

	skipWS := func(pos int) int {
		for pos < len(states) && isIRStateWS(states[pos]) {
			pos++
		}
		return pos
	}

	parseQuoted := func(pos int) (contentStart, contentEnd, next int, ok bool) {
		if pos >= len(states) || states[pos] != '"' {
			return 0, 0, 0, false
		}
		contentStart = pos + 1
		pos = contentStart
		for pos < len(states) {
			if states[pos] == '\\' {
				pos += 2
				continue
			}
			if states[pos] == '"' {
				return contentStart, pos, pos + 1, true
			}
			pos++
		}
		return 0, 0, 0, false
	}

	i2 := skipWS(i)
	if i2 != i {
		needsTidy = true
	}
	i = i2
	if i < len(states) && states[i] == ']' {
		i++
		i2 = skipWS(i)
		if i2 != i {
			needsTidy = true
		}
		if i2 != len(states) {
			return states
		}
		if needsTidy {
			return "[]"
		}
		return states
	}

	for {
		i2 = skipWS(i)
		if i2 != i {
			needsTidy = true
		}
		i = i2

		keyStart, keyEnd, next, ok := parseQuoted(i)
		if !ok {
			return states
		}
		i = next

		i2 = skipWS(i)
		if i2 != i {
			needsTidy = true
		}
		i = i2
		if i >= len(states) || states[i] != '=' {
			return states
		}
		i++

		i2 = skipWS(i)
		if i2 != i {
			needsTidy = true
		}
		i = i2
		if i >= len(states) {
			return states
		}

		valStart, valEnd := 0, 0
		if states[i] == '"' {
			_, _, next, ok = parseQuoted(i)
			if !ok {
				return states
			}
			valStart = i
			valEnd = next
			i = next
		} else {
			tokenStart := i
			for i < len(states) && states[i] != ',' && states[i] != ']' {
				i++
			}
			tokenEnd := i
			rawStart := tokenStart
			rawEnd := tokenEnd
			for rawStart < rawEnd && isIRStateWS(states[rawStart]) {
				rawStart++
			}
			for rawEnd > rawStart && isIRStateWS(states[rawEnd-1]) {
				rawEnd--
			}
			if rawStart != tokenStart || rawEnd != tokenEnd {
				needsTidy = true
			}
			if rawStart == rawEnd {
				return states
			}
			valStart = rawStart
			valEnd = rawEnd
		}

		i2 = skipWS(i)
		if i2 != i {
			needsTidy = true
		}
		i = i2
		if i >= len(states) {
			return states
		}

		if prevKeyStart != -1 && compareStateKey(states, prevKeyStart, prevKeyEnd, keyStart, keyEnd) > 0 {
			sorted = false
		}
		prevKeyStart, prevKeyEnd = keyStart, keyEnd

		items = append(items, stateItem{
			keyStart: keyStart,
			keyEnd:   keyEnd,
			valStart: valStart,
			valEnd:   valEnd,
		})

		if states[i] == ',' {
			i++
			i2 = skipWS(i)
			if i2 != i {
				needsTidy = true
			}
			i = i2
			if i < len(states) && states[i] == ']' {
				needsTidy = true
				i++
				i2 = skipWS(i)
				if i2 != i {
					needsTidy = true
				}
				if i2 != len(states) {
					return states
				}
				break
			}
			continue
		}
		if states[i] == ']' {
			i++
			i2 = skipWS(i)
			if i2 != i {
				needsTidy = true
			}
			if i2 != len(states) {
				return states
			}
			break
		}
		return states
	}

	if !sorted {
		sort.Slice(items, func(i, j int) bool {
			a := items[i]
			b := items[j]
			return compareStateKey(states, a.keyStart, a.keyEnd, b.keyStart, b.keyEnd) < 0
		})
	}

	builder := strings.Builder{}
	builder.Grow(len(states))
	builder.WriteByte('[')
	for idx, it := range items {
		if idx != 0 {
			builder.WriteByte(',')
		}
		builder.WriteByte('"')
		builder.WriteString(states[it.keyStart:it.keyEnd])
		builder.WriteByte('"')
		builder.WriteByte('=')
		builder.WriteString(states[it.valStart:it.valEnd])
	}
	builder.WriteByte(']')
	return builder.String()
}

func compareStateKey(s string, aStart, aEnd, bStart, bEnd int) int {
	aLen := aEnd - aStart
	bLen := bEnd - bStart
	n := aLen
	if bLen < n {
		n = bLen
	}
	for i := 0; i < n; i++ {
		ab := s[aStart+i]
		bb := s[bStart+i]
		if ab < bb {
			return -1
		}
		if ab > bb {
			return 1
		}
	}
	if aLen < bLen {
		return -1
	}
	if aLen > bLen {
		return 1
	}
	return 0
}

func isIRStateWS(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
