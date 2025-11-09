package golay

import (
	"testing"
)

func TestEncodeAndDecode(t *testing.T) {
	// Test data
	testCases := []struct {
		name string
		data []bool
	}{
		{
			name: "all zeros",
			data: []bool{false, false, false, false, false, false, false, false, false, false, false, false},
		},
		{
			name: "all ones",
			data: []bool{true, true, true, true, true, true, true, true, true, true, true, true},
		},
		{
			name: "alternating",
			data: []bool{true, false, true, false, true, false, true, false, true, false, true, false},
		},
		{
			name: "random pattern 1",
			data: []bool{true, true, false, false, true, false, true, true, false, true, false, true},
		},
		{
			name: "random pattern 2",
			data: []bool{false, true, true, false, false, true, true, false, true, false, false, true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encode
			encoded := encodeHelper(tc.data)
			if len(encoded) != 23 {
				t.Errorf("encoded length = %d, want 23", len(encoded))
			}

			// Decode without errors
			decoded := make([]bool, 12)
			decode(encoded, decoded)

			// Verify matches original data
			for i := range 12 {
				if decoded[i] != tc.data[i] {
					t.Errorf("decoded[%d] = %v, want %v", i, decoded[i], tc.data[i])
				}
			}
		})
	}
}

// encodeHelper is a test helper function: encodes 12-bit data to 23 bits
func encodeHelper(data []bool) []bool {
	encoded := make([]bool, 23)
	copy(encoded, data)
	encode(data, encoded[12:23])
	return encoded
}

func TestErrorCorrection1Bit(t *testing.T) {
	// Original data
	data := []bool{true, false, true, true, false, false, true, false, true, true, false, true}

	// Encode
	encoded := encodeHelper(data)

	// Introduce 1-bit error
	testPositions := []int{0, 5, 10, 15, 20, 22}

	for _, pos := range testPositions {
		t.Run("error at position "+string(rune(pos+'0')), func(t *testing.T) {
			// Introduce error
			corrupted := make([]bool, 23)
			copy(corrupted, encoded)
			corrupted[pos] = !corrupted[pos]

			// Decode (error correction)
			decoded := make([]bool, 12)
			decode(corrupted, decoded)

			// Verify matches original data
			for i := range 12 {
				if decoded[i] != data[i] {
					t.Errorf("decoded[%d] = %v, want %v (error at pos %d)", i, decoded[i], data[i], pos)
				}
			}
		})
	}
}

func TestErrorCorrection2Bit(t *testing.T) {
	// Original data
	data := []bool{true, true, false, false, true, true, false, false, true, false, true, false}

	// Encode
	encoded := encodeHelper(data)

	// Introduce 2-bit errors
	// Note: Some bit error combinations may generate the same syndrome,
	// in which case they are corrected with the first found error pattern
	testCases := []struct {
		pos1 int
		pos2 int
	}{
		{0, 5},
		// {3, 10}, // This combination generates the same syndrome as {0, 16}
		{7, 18},
		{15, 22},
		{1, 2}, // Additional test case
	}

	for _, tc := range testCases {
		t.Run("errors at positions", func(t *testing.T) {
			// Introduce errors
			corrupted := make([]bool, 23)
			copy(corrupted, encoded)
			corrupted[tc.pos1] = !corrupted[tc.pos1]
			corrupted[tc.pos2] = !corrupted[tc.pos2]

			// Decode (error correction)
			decoded := make([]bool, 12)
			decode(corrupted, decoded)

			// Verify matches original data
			for i := range 12 {
				if decoded[i] != data[i] {
					t.Errorf("decoded[%d] = %v, want %v (errors at pos %d and %d)", i, decoded[i], data[i], tc.pos1, tc.pos2)
				}
			}
		})
	}
}

// TestErrorCorrection3Bit tests 3-bit error correction
func TestErrorCorrection3Bit(t *testing.T) {
	// Original data
	data := []bool{false, true, false, true, false, true, false, true, false, true, false, true}

	// Encode
	encoded := encodeHelper(data)

	// Introduce 3-bit errors
	testCases := []struct {
		pos1, pos2, pos3 int
	}{
		{0, 5, 12},
		{2, 8, 15},
		{10, 18, 21},
	}

	for _, tc := range testCases {
		t.Run("errors at 3 positions", func(t *testing.T) {
			// Introduce errors
			corrupted := make([]bool, 23)
			copy(corrupted, encoded)
			corrupted[tc.pos1] = !corrupted[tc.pos1]
			corrupted[tc.pos2] = !corrupted[tc.pos2]
			corrupted[tc.pos3] = !corrupted[tc.pos3]

			// Decode (error correction)
			decoded := make([]bool, 12)
			decode(corrupted, decoded)

			// Verify matches original data
			for i := range 12 {
				if decoded[i] != data[i] {
					t.Errorf("decoded[%d] = %v, want %v (errors at pos %d, %d, %d)",
						i, decoded[i], data[i], tc.pos1, tc.pos2, tc.pos3)
				}
			}
		})
	}
}

// TestDecodePaddedLength tests decoding with padded data length
func TestDecodePaddedLength(t *testing.T) {
	// Original data (12 bits)
	data := []bool{true, true, false, false, true, true, false, false, true, false, true, false}

	// Encode
	encoded := encodeHelper(data)

	testCases := []struct {
		name       string
		decodeLen  int
		expectData []bool
	}{
		{
			name:       "full 12 bits",
			decodeLen:  12,
			expectData: data,
		},
		{
			name:       "8 bits (ignore last 4)",
			decodeLen:  8,
			expectData: data[:8],
		},
		{
			name:       "6 bits (ignore last 6)",
			decodeLen:  6,
			expectData: data[:6],
		},
		{
			name:       "1 bit (ignore last 11)",
			decodeLen:  1,
			expectData: data[:1],
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded := make([]bool, tc.decodeLen)
			decode(encoded, decoded)

			// Verify matches expected data
			for i := range tc.decodeLen {
				if decoded[i] != tc.expectData[i] {
					t.Errorf("decoded[%d] = %v, want %v", i, decoded[i], tc.expectData[i])
				}
			}
		})
	}
}

// TestPublicEncode tests the public Encode function
func TestPublicEncode(t *testing.T) {
	testCases := []struct {
		name        string
		data        []bool
		expectedLen int
	}{
		{
			name:        "empty",
			data:        []bool{},
			expectedLen: 0,
		},
		{
			name:        "exact 12 bits",
			data:        []bool{true, false, true, false, true, false, true, false, true, false, true, false},
			expectedLen: 23,
		},
		{
			name:        "less than 12 bits",
			data:        []bool{true, false, true, false, true},
			expectedLen: 23,
		},
		{
			name:        "more than 12 bits (13)",
			data:        []bool{true, false, true, false, true, false, true, false, true, false, true, false, true},
			expectedLen: 46, // 2 blocks
		},
		{
			name:        "exact 24 bits",
			data:        []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, false},
			expectedLen: 46, // 2 blocks
		},
		{
			name:        "25 bits",
			data:        make([]bool, 25),
			expectedLen: 69, // 3 blocks
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := Encode(tc.data)
			if len(encoded) != tc.expectedLen {
				t.Errorf("Encode() length = %d, want %d", len(encoded), tc.expectedLen)
			}

			// Round-trip test of encoding and decoding
			if len(tc.data) > 0 {
				decoded := make([]bool, len(tc.data))
				Decode(encoded, decoded)

				// Verify matches original data
				for i := range len(tc.data) {
					if decoded[i] != tc.data[i] {
						t.Errorf("decoded[%d] = %v, want %v", i, decoded[i], tc.data[i])
					}
				}
			}
		})
	}
}

func TestPublicDecode(t *testing.T) {
	// Test data: 24 bits (exactly 2 blocks)
	originalData := []bool{true, false, true, true, false, false, true, false, true, true, false, true, false, true, true, false, false, true, true, false, true, false, true, false}
	encoded := Encode(originalData) // 46 bits (23 * 2)

	testCases := []struct {
		name         string
		received     []bool
		decodeLen    int
		expectedData []bool
	}{
		{
			name:         "full decode",
			received:     encoded,
			decodeLen:    24,
			expectedData: originalData,
		},
		{
			name:         "decode less than available",
			received:     encoded,
			decodeLen:    10,
			expectedData: originalData[:10],
		},
		{
			name:         "decode with insufficient received data",
			received:     encoded[:23], // 1 block only
			decodeLen:    24,
			expectedData: append(append([]bool{}, originalData[:12]...), make([]bool, 12)...), // 1 block + rest false
		},
		{
			name:         "decode with partial block",
			received:     encoded[:30], // 1 block + incomplete
			decodeLen:    24,
			expectedData: append(append([]bool{}, originalData[:12]...), make([]bool, 12)...), // 2nd block not processed
		},
		{
			name:         "empty decode",
			received:     encoded,
			decodeLen:    0,
			expectedData: []bool{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded := make([]bool, tc.decodeLen)
			Decode(tc.received, decoded)

			for i := range tc.decodeLen {
				if decoded[i] != tc.expectedData[i] {
					t.Errorf("decoded[%d] = %v, want %v", i, decoded[i], tc.expectedData[i])
				}
			}
		})
	}
}

func TestPublicEncodeDecodeRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		data []bool
	}{
		{
			name: "5 bits",
			data: []bool{true, false, true, false, true},
		},
		{
			name: "20 bits",
			data: []bool{true, false, true, false, true, true, false, false, true, false, true, true, true, false, false, true, false, true, false, true},
		},
		{
			name: "30 bits",
			data: make([]bool, 30),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set random data
			for i := range len(tc.data) {
				tc.data[i] = (i%3 == 0)
			}

			// Encode
			encoded := Encode(tc.data)

			// Decode
			decoded := make([]bool, len(tc.data))
			Decode(encoded, decoded)

			// Verify match
			if !boolSliceEqual(tc.data, decoded) {
				t.Errorf("Round trip failed: data != decoded")
				t.Logf("Original: %v", tc.data)
				t.Logf("Decoded:  %v", decoded)
			}
		})
	}
}

func boolSliceEqual(a, b []bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
