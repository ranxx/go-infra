package utils

import (
	"crypto/md5"
	"encoding/hex"
)

const (
	hashBase36 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Hash hash
func Hash(str ...string) string {
	hash := md5.New()

	for _, v := range str {
		hash.Write([]byte(v))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// HashCode generates a base36 code (A-Z, 0-9) of the specified length from input strings.
// Same inputs always produce the same output.
func HashCode(length int, str ...string) string {
	hash := md5.New()
	for _, v := range str {
		hash.Write([]byte(v))
	}
	b := hash.Sum(nil)

	// Pack first 8 bytes (64 bits) and extract 60 bits (10 chars × 6 bits)
	var val uint64
	for i := 0; i < 8; i++ {
		val = (val << 8) | uint64(b[i])
	}

	result := make([]byte, length)
	for i := int(length) - 1; i >= 0; i-- {
		result[i] = hashBase36[(val&0x3F)%36]
		val >>= 6
	}

	return string(result)
}
