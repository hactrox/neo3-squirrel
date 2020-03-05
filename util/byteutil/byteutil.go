package byteutil

// ReverseBytes reverses the given bytes
func ReverseBytes(raw []byte) []byte {
	if len(raw) == 0 {
		return raw
	}

	reversed := make([]byte, len(raw))

	for i := len(raw) - 1; i >= 0; i-- {
		reversed[len(raw)-i-1] = raw[i]
	}

	return reversed
}
