package report

import (
	"fmt"
	"reflect"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer"
)

func DiffOpaqueMaps(path string, map1, map2 deployer.OpaqueMap) []string {
	differences := []string{}

	// Check keys in map1
	for key, val1 := range map1 {
		if val2, exists := map2[key]; !exists {
			differences = append(differences, fmt.Sprintf("%s.%s: exists in first map but not in second (value: %v)", path, key, val1))
		} else {
			differences = append(differences, diffValues(path+"."+key, val1, val2)...)
		}
	}

	// Check for keys in map2 that aren't in map1
	for key, val2 := range map2 {
		if _, exists := map1[key]; !exists {
			differences = append(differences, fmt.Sprintf("%s.%s: exists in second map but not in first (value: %v)", path, key, val2))
		}
	}

	return differences
}

func diffValues(path string, val1, val2 interface{}) []string {
	var differences []string

	switch v1 := val1.(type) {
	case map[string]interface{}:
		if v2, ok := val2.(map[string]interface{}); ok {
			// Convert to OpaqueMapping and recurse
			om1 := deployer.OpaqueMap(v1)
			om2 := deployer.OpaqueMap(v2)
			differences = append(differences, DiffOpaqueMaps(path, om1, om2)...)
		} else {
			differences = append(differences, fmt.Sprintf("%s: type mismatch (map vs %T)", path, val2))
		}
	case []interface{}:
		if v2, ok := val2.([]interface{}); ok {
			if len(v1) != len(v2) {
				differences = append(differences, fmt.Sprintf("%s: array length mismatch (%d vs %d)", path, len(v1), len(v2)))
			} else {
				for i := range v1 {
					differences = append(differences, diffValues(fmt.Sprintf("%s[%d]", path, i), v1[i], v2[i])...)
				}
			}
		} else {
			differences = append(differences, fmt.Sprintf("%s: type mismatch (array vs %T)", path, val2))
		}
	default:
		if !reflect.DeepEqual(val1, val2) {
			differences = append(differences, fmt.Sprintf("%s: value mismatch (%v vs %v)", path, val1, val2))
		}
	}

	return differences
}
