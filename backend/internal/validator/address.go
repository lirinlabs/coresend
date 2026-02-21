package validator

func IsValidHexAddress(s string) bool {
	if len(s) != 40 {
		return false
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}

	return true
}
