package opaque_map

type (
	OpaqueMap   map[string]any
	OpaqueState OpaqueMap
)

// UseInts converts all float64 values without fractional parts to int64 values in a map
// so that they are properly marshaled to TOML
func UseInts(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case float64:
			// If the float has no fractional part, convert to int
			if val == float64(int64(val)) {
				m[k] = int64(val)
			}
		case map[string]any:
			// Recursively process nested maps
			UseInts(val)
		case []any:
			// Process arrays
			for i, item := range val {
				if fItem, ok := item.(float64); ok && fItem == float64(int64(fItem)) {
					val[i] = int64(fItem)
				} else if mapItem, ok := item.(map[string]any); ok {
					UseInts(mapItem)
				}
			}
		}
	}
}
