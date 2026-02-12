package ir_states

import (
	"strconv"
)

// CastToString converts all values to quoted string values.
// bool true/false become "1"/"0", and int values become decimal text.
func CastToString(states string) string {
	if !IsValid(states) {
		return states
	}

	entries := map[string]string{}
	KeyValues(states)(func(key string, value any) bool {
		strValue, ok := castValueToString(value)
		if !ok {
			return false
		}
		entries[key] = strconv.Quote(strValue)
		return true
	})
	return buildSortedStates(entries)
}

// FuzzyCompare compares two states texts by key/value map.
// It returns:
// same: keys present in both with same value,
// different: keys present in both with different values,
// missing: keys missing in s1 but present in s2,
// extra: keys present in s1 but missing in s2.
func FuzzyCompare(s1, s2 string) (same, different, missing, extra int) {
	m1, ok1 := toComparableMap(s1)
	m2, ok2 := toComparableMap(s2)
	if !ok1 || !ok2 {
		return 0, 0, 0, 0
	}

	for k, v2 := range m2 {
		v1, found := m1[k]
		if !found {
			missing++
			continue
		}
		if v1 == v2 {
			same++
		} else {
			different++
		}
	}

	for k := range m1 {
		if _, found := m2[k]; !found {
			extra++
		}
	}

	return same, different, missing, extra
}

func toComparableMap(states string) (map[string]string, bool) {
	if !IsValid(states) {
		return nil, false
	}
	out := map[string]string{}
	ok := true
	KeyValues(states)(func(key string, value any) bool {
		s, canCast := castValueToString(value)
		if !canCast {
			ok = false
			return false
		}
		out[key] = s
		return true
	})
	return out, ok
}

func castValueToString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case int8:
		if v == 0 {
			return "0", true
		}
		if v == 1 {
			return "1", true
		}
		return "", false
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	default:
		return "", false
	}
}
