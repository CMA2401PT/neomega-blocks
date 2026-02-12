package ir_states

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func Set(states string, key string, value any) string {
	if key == "" || !IsValid(states) {
		return states
	}

	rawValue, ok := encodeSetValue(value)
	if !ok {
		return states
	}

	entries, ok := collectCanonicalEntries(states)
	if !ok {
		return states
	}
	entries[key] = rawValue
	return buildSortedStates(entries)
}

func Delete(states string, key string) string {
	if key == "" || !IsValid(states) {
		return states
	}

	entries, ok := collectCanonicalEntries(states)
	if !ok {
		return states
	}
	if _, found := entries[key]; !found {
		return states
	}
	delete(entries, key)
	return buildSortedStates(entries)
}

func collectCanonicalEntries(states string) (map[string]string, bool) {
	entries := map[string]string{}
	ok := true
	KeyValues(states)(func(key string, value any) bool {
		raw, canEncode := encodeSetValue(value)
		if !canEncode {
			ok = false
			return false
		}
		entries[key] = raw
		return true
	})
	return entries, ok
}

func buildSortedStates(entries map[string]string) string {
	if len(entries) == 0 {
		return "[]"
	}

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	builder := &strings.Builder{}
	builder.WriteByte('[')
	for i, k := range keys {
		if i != 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.Quote(k))
		builder.WriteByte('=')
		builder.WriteString(entries[k])
	}
	builder.WriteByte(']')
	return builder.String()
}

func encodeSetValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return strconv.Quote(v), true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	case int8:
		if v == 0 {
			return "false", true
		}
		if v == 1 {
			return "true", true
		}
		return "", false
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return "", false
		}
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		if v < math.MinInt32 || v > math.MaxInt32 {
			return "", false
		}
		return strconv.FormatInt(v, 10), true
	case uint:
		if v > math.MaxInt32 {
			return "", false
		}
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		if v > math.MaxInt32 {
			return "", false
		}
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		if v > math.MaxInt32 {
			return "", false
		}
		return strconv.FormatUint(v, 10), true
	default:
		panic(fmt.Errorf("unknown value type %T:%v", value, value))
	}
}
