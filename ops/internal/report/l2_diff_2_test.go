package report

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
	"github.com/stretchr/testify/require"
)

func TestDiffOpaqueMaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		map1     deployer.OpaqueMap
		map2     deployer.OpaqueMap
		expected []string
	}{
		{
			name: "identical maps",
			map1: deployer.OpaqueMap{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			map2: deployer.OpaqueMap{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			expected: []string{},
		},
		{
			name: "different values",
			map1: deployer.OpaqueMap{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			map2: deployer.OpaqueMap{
				"key1": "different",
				"key2": 42,
				"key3": false,
			},
			expected: []string{
				"path.key1: value mismatch (value1 vs different)",
				"path.key3: value mismatch (true vs false)",
			},
		},
		{
			name: "missing keys",
			map1: deployer.OpaqueMap{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			map2: deployer.OpaqueMap{
				"key1": "value1",
				"key4": "new",
			},
			expected: []string{
				"path.key2: exists in first map but not in second (value: 42)",
				"path.key3: exists in first map but not in second (value: true)",
				"path.key4: exists in second map but not in first (value: new)",
			},
		},
		{
			name: "nested maps",
			map1: deployer.OpaqueMap{
				"key1": map[string]interface{}{
					"nested1": "value1",
					"nested2": 42,
				},
			},
			map2: deployer.OpaqueMap{
				"key1": map[string]interface{}{
					"nested1": "different",
					"nested3": true,
				},
			},
			expected: []string{
				"path.key1.nested1: value mismatch (value1 vs different)",
				"path.key1.nested2: exists in first map but not in second (value: 42)",
				"path.key1.nested3: exists in second map but not in first (value: true)",
			},
		},
		{
			name: "deeply nested fields",
			map1: deployer.OpaqueMap{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"level4": map[string]interface{}{
								"value": 42,
								"array": []interface{}{1, 2, 3},
								"flag":  true,
							},
							"sibling": "unchanged",
						},
						"level3b": "unchanged",
					},
				},
			},
			map2: deployer.OpaqueMap{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": map[string]interface{}{
							"level4": map[string]interface{}{
								"value": 100,
								"array": []interface{}{1, 5, 3},
								"flag":  false,
								"new":   "added",
							},
							"sibling": "unchanged",
						},
						"level3b": "unchanged",
					},
				},
			},
			expected: []string{
				"path.level1.level2.level3.level4.value: value mismatch (42 vs 100)",
				"path.level1.level2.level3.level4.array[1]: value mismatch (2 vs 5)",
				"path.level1.level2.level3.level4.flag: value mismatch (true vs false)",
				"path.level1.level2.level3.level4.new: exists in second map but not in first (value: added)",
			},
		},
		{
			name: "arrays",
			map1: deployer.OpaqueMap{
				"array": []interface{}{1, 2, 3},
			},
			map2: deployer.OpaqueMap{
				"array": []interface{}{1, 4, 3},
			},
			expected: []string{
				"path.array[1]: value mismatch (2 vs 4)",
			},
		},
		{
			name: "array length mismatch",
			map1: deployer.OpaqueMap{
				"array": []interface{}{1, 2, 3},
			},
			map2: deployer.OpaqueMap{
				"array": []interface{}{1, 2},
			},
			expected: []string{
				"path.array: array length mismatch (3 vs 2)",
			},
		},
		{
			name: "type mismatch",
			map1: deployer.OpaqueMap{
				"key1": map[string]interface{}{"nested": "value"},
				"key2": []interface{}{1, 2, 3},
			},
			map2: deployer.OpaqueMap{
				"key1": "not a map",
				"key2": "not an array",
			},
			expected: []string{
				"path.key1: type mismatch (map vs string)",
				"path.key2: type mismatch (array vs string)",
			},
		},
		{
			name:     "empty maps",
			map1:     deployer.OpaqueMap{},
			map2:     deployer.OpaqueMap{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := DiffOpaqueMaps("path", tt.map1, tt.map2)
			require.ElementsMatch(t, tt.expected, diffs)
		})
	}
}
