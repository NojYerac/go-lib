package audit

import "reflect"

type Change struct {
	Before any `json:"before,omitempty"`
	After  any `json:"after,omitempty"`
}

func CompactDiff(before, after map[string]any) map[string]Change {
	if len(before) == 0 && len(after) == 0 {
		return nil
	}

	diff := make(map[string]Change)
	for key, beforeValue := range before {
		afterValue, ok := after[key]
		if !ok {
			diff[key] = Change{Before: beforeValue}
			continue
		}

		if !reflect.DeepEqual(beforeValue, afterValue) {
			diff[key] = Change{Before: beforeValue, After: afterValue}
		}
	}

	for key, afterValue := range after {
		if _, ok := before[key]; ok {
			continue
		}
		diff[key] = Change{After: afterValue}
	}

	if len(diff) == 0 {
		return nil
	}

	return diff
}
