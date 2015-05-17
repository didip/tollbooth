package libstring

func FlattenMapSliceString(mapSliceString map[string][]string, prefix string) []string {
	result := make([]string, 0)

	for key, slice := range mapSliceString {
		for _, item := range slice {
			result = append(result, prefix+":"+key+":"+item)
		}
	}
	return result
}
