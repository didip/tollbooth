// Package libstring provides string related functions.
package libstring

func FlattenMapSliceString(mapSliceString map[string][]string, prefix string, separator string) []string {
	result := make([]string, 0)

	if separator == "" {
		separator = ":"
	}

	for key, slice := range mapSliceString {
		for _, item := range slice {
			result = append(result, prefix+separator+key+separator+item)
		}
	}
	return result
}
