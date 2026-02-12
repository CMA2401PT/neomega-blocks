package ir_states

import (
	"strconv"
	"strings"
)

// KeyValueStrings iterates key/value pairs in IRStates and supports:
// for k, v := range s.KeyValueStrings() { ... }
// key is returned without quotes; value is returned exactly as in IRStates.
func KeyValueStrings(states string) func(func(string, string) bool) {
	return func(yield func(string, string) bool) {
		if len(states) < 2 || states[0] != '[' || states[len(states)-1] != ']' {
			return
		}
		if states == "[]" {
			return
		}

		i := 1
		last := len(states) - 1
		for {
			for i < last && isIRStateWS(states[i]) {
				i++
			}
			if i >= last || states[i] != '"' {
				return
			}
			keyStart := i + 1
			i = keyStart
			for i < last && states[i] != '"' {
				i++
			}
			if i >= last {
				return
			}
			keyEnd := i
			i++

			for i < last && isIRStateWS(states[i]) {
				i++
			}
			if i >= last || states[i] != '=' {
				return
			}
			i++
			for i < last && isIRStateWS(states[i]) {
				i++
			}
			if i >= len(states) {
				return
			}

			valStart := i
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
					return
				}
				i++
			} else {
				for i < last && states[i] != ',' && states[i] != ']' {
					i++
				}
				if valStart == i {
					return
				}
			}
			valEnd := i

			if !yield(states[keyStart:keyEnd], states[valStart:valEnd]) {
				return
			}

			for i < last && isIRStateWS(states[i]) {
				i++
			}
			if i >= len(states) {
				return
			}
			switch states[i] {
			case ',':
				i++
				continue
			case ']':
				return
			default:
				return
			}
		}
	}
}
func KeyValues(states string) func(func(string, any) bool) {
	return func(yield func(string, any) bool) {
		KeyValueStrings(states)(func(key, raw string) bool {
			if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
				v, err := strconv.Unquote(raw)
				if err != nil {
					return false
				}
				return yield(key, v)
			}

			trimmed := strings.TrimSpace(raw)
			switch trimmed {
			case "true":
				return yield(key, int8(1))
			case "false":
				return yield(key, int8(0))
			}

			n, err := strconv.ParseInt(trimmed, 10, 32)
			if err != nil {
				return false
			}
			return yield(key, int32(n))
		})
	}
}
