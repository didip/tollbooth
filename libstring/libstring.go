package libstring

func StringInSlice(list []string, needle string) bool {
	for _, b := range list {
		if b == needle {
			return true
		}
	}
	return false
}
