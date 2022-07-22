package core

func SliceContains[T comparable](arr []T, x T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}

func IsDecimal(a float64) bool {
	if a == float64(int64(a)) {
		return true
	} else {
		return false
	}
}
