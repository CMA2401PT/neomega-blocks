package ir_states

import (
	"strconv"
	"strings"
)

// Canon canonicalizes states against std:
//   - keys not in std are removed,
//   - keys missing from states are filled from std defaults,
//   - keys present in both keep states value when type matches std,
//     or after a best-effort conversion; otherwise fallback to std default.
//
// Result is returned in canonical sorted form.
func Canon(states string, std string) string {
	if !IsValid(states) || !IsValid(std) {
		return states
	}

	stateValues := map[string]any{}
	KeyValues(states)(func(key string, value any) bool {
		stateValues[key] = value
		return true
	})
	stdValues := map[string]any{}
	KeyValues(std)(func(key string, value any) bool {
		stdValues[key] = value
		return true
	})

	out := make(map[string]string, len(stdValues))
	for k, stdValue := range stdValues {
		valueToUse := stdValue
		if inValue, found := stateValues[k]; found {
			if coerced, ok := canonCoerceToType(inValue, stdValue); ok {
				valueToUse = coerced
			} else {
				valueToUse = inValue
			}
		}
		raw, ok := encodeSetValue(valueToUse)
		if !ok {
			return states
		}
		out[k] = raw
	}
	return buildSortedStates(out)
}

func canonCoerceToType(inValue any, stdValue any) (any, bool) {
	switch stdValue.(type) {
	case int32:
		switch v := inValue.(type) {
		case int32:
			return v, true
		case int8:
			if v == 0 || v == 1 {
				return int32(v), true
			}
			return nil, false
		case string:
			s := strings.TrimSpace(strings.ToLower(v))
			switch s {
			case "true":
				return int32(1), true
			case "false":
				return int32(0), true
			}
			n, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return nil, false
			}
			return int32(n), true
		default:
			return nil, false
		}
	case int8:
		switch v := inValue.(type) {
		case int8:
			if v == 0 || v == 1 {
				return v, true
			}
			return nil, false
		case int32:
			if v == 0 || v == 1 {
				return int8(v), true
			}
			return nil, false
		case string:
			s := strings.TrimSpace(strings.ToLower(v))
			switch s {
			case "true":
				return int8(1), true
			case "false":
				return int8(0), true
			}
			n, err := strconv.ParseInt(s, 10, 32)
			if err != nil || (n != 0 && n != 1) {
				return nil, false
			}
			return int8(n), true
		default:
			return nil, false
		}
	case string:
		switch v := inValue.(type) {
		case string:
			return v, true
		case int32:
			return strconv.FormatInt(int64(v), 10), true
		case int8:
			if v == 0 {
				return "0", true
			}
			if v == 1 {
				return "1", true
			}
			return nil, false
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}
