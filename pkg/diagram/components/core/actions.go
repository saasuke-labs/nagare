package core

// ActionMapFromAny casts a generic any value to the action map type used by
// diagram components. It returns an empty map when the cast fails.
func ActionMapFromAny(v any) map[string][]map[string]any {
	out := make(map[string][]map[string]any)
	if v == nil {
		return out
	}

	typed, ok := v.(map[string][]map[string]any)
	if !ok {
		return out
	}

	for key, entries := range typed {
		copied := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			item := make(map[string]any, len(entry))
			for k, val := range entry {
				item[k] = val
			}
			copied = append(copied, item)
		}
		out[key] = copied
	}

	return out
}
