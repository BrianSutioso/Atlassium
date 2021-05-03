package utils

// RevStrArr (ReverseStringArray) reverses
// the order of an array of strings in place.
// Inputs:
// a []string the array of strings to be
// reversed in place
func RevStrArr(a []string) {
	last := len(a) - 1
	for i := 0; i < len(a)/2; i++ {
		a[i], a[last-i] = a[last-i], a[i]
	}
}

// CalcPOWD (CalculateProofOfWorkDifficulty) takes in
// a hardness level a.k.a the number of leading zeros
// that the difficulty target should have. From this
// hardness level, it returns the initial difficulty
// target as a hex string.
// Inputs:
// powdNumZeros int the number of leading zeros of
// the target difficulty hex string
// Returns:
// string the target difficulty represented as a hex
// string. Should be 64 characters long since it should
// stay consistent with 32 byte hashing (SHA-256).
func CalcPOWD(powdNumZeros int) string {
	if powdNumZeros <= -1 || powdNumZeros >= 30 {
		powdNumZeros = 3
	}
	ret := ""
	for i := 0; i < 64; i++ {
		if i != powdNumZeros {
			ret += "0"
		} else {
			ret += "1"
		}
	}
	return ret
}

// InSlice tells whether a given string is
// in an array of strings
// Inputs:
// a []string the array of strings to check
// membership against
// val string the value that may be in the
// array
// Returns:
// bool True if the val is in the array. False
// otherwise
func InSlice(a []string, val string) bool {
	for _, v := range a {
		if v == val {
			return true
		}
	}
	return false
}
