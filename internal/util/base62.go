package util

import (
	"errors"
	"strings"
)

// base62Chars is the character set used for Base62 encoding
// Includes 0-9, a-z, A-Z (62 characters total)
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeBase62 converts a uint64 number to a Base62 string
// Example: 62 -> "10", 12345 -> "3D7"
func EncodeBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var result strings.Builder
	base := uint64(62)

	// Build string in reverse order
	for num > 0 {
		result.WriteByte(base62Chars[num%base])
		num /= base
	}

	// Reverse the string since we built it backwards
	s := result.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// DecodeBase62 converts a Base62 string back to a uint64 number
// Returns an error if the string contains invalid characters
func DecodeBase62(s string) (uint64, error) {
	var result uint64
	base := uint64(62)

	for _, char := range s {
		idx := strings.IndexRune(base62Chars, char)
		if idx == -1 {
			return 0, errors.New("invalid base62 character")
		}
		result = result*base + uint64(idx)
	}
	return result, nil
}
