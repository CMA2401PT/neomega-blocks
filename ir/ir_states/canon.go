package ir_states

// Canon canonicalizes states against std:
// - keys not in std are removed,
// - keys missing from states are filled from std defaults,
// - keys present in both keep states value.
// Result is returned in canonical sorted form.
func Canon(states string, std string) string {
	if !IsValid(states) || !IsValid(std) {
		return states
	}

	stateEntries, ok := collectCanonicalEntries(states)
	if !ok {
		return states
	}
	stdEntries, ok := collectCanonicalEntries(std)
	if !ok {
		return states
	}

	out := make(map[string]string, len(stdEntries))
	for k, defaultValue := range stdEntries {
		if v, found := stateEntries[k]; found {
			out[k] = v
		} else {
			out[k] = defaultValue
		}
	}
	return buildSortedStates(out)
}
