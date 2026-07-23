package util

import "testing"

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{61, "Z"},
		{62, "10"},
		{12345, "3D7"},
		{1000000, "4C92"},
	}

	for _, test := range tests {
		result := EncodeBase62(test.input)
		if result != test.expected {
			t.Errorf("EncodeBase62(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0", 0},
		{"Z", 61},
		{"10", 62},
		{"3D7", 12345},
		{"4C92", 1000000},
	}

	for _, test := range tests {
		result, err := DecodeBase62(test.input)
		if err != nil {
			t.Fatalf("DecodeBase62(%s) returned error: %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("DecodeBase62(%s) = %d; want %d", test.input, result, test.expected)
		}
	}
}

func TestDecodeBase62Invalid(t *testing.T) {
	_, err := DecodeBase62("!!!")
	if err == nil {
		t.Error("Expected error for invalid base62 string, got nil")
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	numbers := []uint64{0, 1, 61, 62, 100, 12345, 999999, 123456789}

	for _, num := range numbers {
		encoded := EncodeBase62(num)
		decoded, err := DecodeBase62(encoded)
		if err != nil {
			t.Fatalf("Round-trip failed for %d: %v", num, err)
		}
		if decoded != num {
			t.Errorf("Round-trip failed: %d -> %s -> %d", num, encoded, decoded)
		}
	}
}
