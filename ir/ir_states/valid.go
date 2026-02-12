package ir_states

import "strconv"

// IsValid reports whether states is syntactically valid IR states text.
// Unlike IsTidy, it allows extra whitespace and unsorted keys.
func IsValid(states string) bool {
	if len(states) < 2 {
		return false
	}

	parseQuoted := func(pos int) (next int, ok bool) {
		if pos >= len(states) || states[pos] != '"' {
			return 0, false
		}
		start := pos
		pos++
		for pos < len(states) {
			if states[pos] == '\\' {
				pos += 2
				continue
			}
			if states[pos] == '"' {
				if _, err := strconv.Unquote(states[start : pos+1]); err != nil {
					return 0, false
				}
				return pos + 1, true
			}
			pos++
		}
		return 0, false
	}

	skipWS := func(pos int) int {
		for pos < len(states) && isIRStateWS(states[pos]) {
			pos++
		}
		return pos
	}

	i := skipWS(0)
	if i >= len(states) || states[i] != '[' {
		return false
	}
	i++
	i = skipWS(i)

	if i < len(states) && states[i] == ']' {
		i++
		return skipWS(i) == len(states)
	}

	for {
		i = skipWS(i)
		next, ok := parseQuoted(i)
		if !ok {
			return false
		}
		i = skipWS(next)
		if i >= len(states) || states[i] != '=' {
			return false
		}

		i++
		i = skipWS(i)
		if i >= len(states) {
			return false
		}

		if states[i] == '"' {
			next, ok = parseQuoted(i)
			if !ok {
				return false
			}
			i = next
		} else {
			start := i
			for i < len(states) && states[i] != ',' && states[i] != ']' {
				i++
			}
			end := i
			for start < end && isIRStateWS(states[start]) {
				start++
			}
			for end > start && isIRStateWS(states[end-1]) {
				end--
			}
			if start == end {
				return false
			}
			token := states[start:end]
			if token != "true" && token != "false" {
				if _, err := strconv.ParseInt(token, 10, 32); err != nil {
					return false
				}
			}
		}

		i = skipWS(i)
		if i >= len(states) {
			return false
		}
		if states[i] == ',' {
			i++
			i = skipWS(i)
			// allow trailing comma: ["a"=1,]
			if i < len(states) && states[i] == ']' {
				i++
				return skipWS(i) == len(states)
			}
			continue
		}
		if states[i] == ']' {
			i++
			return skipWS(i) == len(states)
		}
		return false
	}
}
