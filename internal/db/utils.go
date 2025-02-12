package db

import "strings"

func unflattenFields(fields map[string]interface{}) {
	for k, col := range fields {
		if strings.Contains(k, ".") { // checks if the field is flattened
			parts := strings.Split(k, ".")
			current := fields
			for i := 0; i < len(parts)-1; i++ {
				subkey := parts[i]
				if next, exits := current[subkey]; exits {
					if nextMap, ok := next.(map[string]interface{}); ok {
						current = nextMap
					} else {
						newMap := make(map[string]interface{})
						current[subkey] = newMap
						current = newMap
					}
				} else {
					newMap := make(map[string]interface{})
					current[subkey] = newMap
					current = newMap
				}
			}
			current[parts[len(parts)-1]] = col
			delete(fields, k)
		}
	}
}
