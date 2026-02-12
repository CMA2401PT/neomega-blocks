package ir_states

// Update applies fn to each key/value entry and returns canonical states text.
// fn receives typed values (string, int32, int8(0/1)) and returns the updated pair.
func Update(states string, fn func(string, any) (string, any)) string {
	if !IsValid(states) || fn == nil {
		return states
	}

	updated := map[string]string{}
	ok := true
	KeyValues(states)(func(key string, value any) bool {
		nextKey, nextValue := fn(key, value)
		if nextKey == "" {
			ok = false
			return false
		}
		raw, canEncode := encodeSetValue(nextValue)
		if !canEncode {
			ok = false
			return false
		}
		updated[nextKey] = raw
		return true
	})
	if !ok {
		return states
	}
	return buildSortedStates(updated)
}
