package ir_states

// Get looks up the first matched key in states and returns its typed value.
func Get(states string, key string) (value any, found bool) {
	KeyValues(states)(func(k string, v any) bool {
		if k == key {
			value = v
			found = true
			return false
		}
		return true
	})
	return value, found
}
