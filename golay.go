package golay

// Generator matrix G parity check portion (12x11 matrix represented as 1D array)
// Using standard Golay(23,12) generator matrix
// G[row][col] = g[row*11 + col]
var g = []bool{
	true, false, true, false, true, true, true, false, false, false, true,
	true, true, false, true, false, true, true, true, false, false, false,
	true, true, true, false, true, false, true, true, true, false, false,
	true, true, true, true, false, true, false, true, true, true, false,
	true, true, true, true, true, false, true, false, true, true, true,
	false, true, true, true, true, true, false, true, false, true, true,
	false, false, true, true, true, true, true, false, true, false, true,
	true, false, false, true, true, true, true, true, false, true, false,
	false, true, false, false, true, true, true, true, true, false, true,
	true, false, true, false, false, true, true, true, true, true, false,
	false, true, false, true, false, false, true, true, true, true, true,
	true, false, true, true, true, false, false, true, true, true, true,
}

// Parity check matrix H transposed (23x11 matrix represented as 1D array)
// H[row][col] = h[row*11 + col]
var h = []bool{
	true, false, true, false, true, true, true, false, false, false, true,
	true, true, false, true, false, true, true, true, false, false, false,
	true, true, true, false, true, false, true, true, true, false, false,
	true, true, true, true, false, true, false, true, true, true, false,
	true, true, true, true, true, false, true, false, true, true, true,
	false, true, true, true, true, true, false, true, false, true, true,
	false, false, true, true, true, true, true, false, true, false, true,
	true, false, false, true, true, true, true, true, false, true, false,
	false, true, false, false, true, true, true, true, true, false, true,
	true, false, true, false, false, true, true, true, true, true, false,
	false, true, false, true, false, false, true, true, true, true, true,
	true, false, true, true, true, false, false, true, true, true, true,
	true, false, false, false, false, false, false, false, false, false, false,
	false, true, false, false, false, false, false, false, false, false, false,
	false, false, true, false, false, false, false, false, false, false, false,
	false, false, false, true, false, false, false, false, false, false, false,
	false, false, false, false, true, false, false, false, false, false, false,
	false, false, false, false, false, true, false, false, false, false, false,
	false, false, false, false, false, false, true, false, false, false, false,
	false, false, false, false, false, false, false, true, false, false, false,
	false, false, false, false, false, false, false, false, true, false, false,
	false, false, false, false, false, false, false, false, false, true, false,
	false, false, false, false, false, false, false, false, false, false, true,
}

// Encode takes data of arbitrary length and performs Golay(23,12) encoding.
// If data length is not a multiple of 12, it pads with false.
// Returns encoded data in 23-bit blocks.
func Encode(data []bool) []bool {
	if len(data) == 0 {
		return []bool{}
	}

	// Calculate number of blocks needed to divide into 12-bit units
	numBlocks := (len(data) + 11) / 12
	result := make([]bool, numBlocks*23)

	for i := range numBlocks {
		// Extract 12 bits at a time
		start := i * 12
		end := min(start+12, len(data))

		// Encoding: 12 data bits + 11 parity bits = 23 bits
		resultStart := i * 23

		// Copy data portion
		copy(result[resultStart:resultStart+12], data[start:end])
		// Remaining bits are automatically initialized to false

		// Calculate parity and write directly to result
		encode(result[resultStart:resultStart+12], result[resultStart+12:resultStart+23])
	}

	return result
}

// Decode decodes encoded data.
// received: encoded data (in 23-bit units)
// data: slice to store decoded results
//
// Determines required number of blocks based on data length:
// - If received is insufficient, decodes as much as possible
// - If received has excess, decodes only what's needed for data length
func Decode(received []bool, data []bool) {
	if len(data) == 0 {
		return
	}

	// Calculate number of blocks needed from data length
	numBlocksNeeded := (len(data) + 11) / 12

	// Calculate number of blocks available from received
	numBlocksAvailable := len(received) / 23

	// Actual number of blocks to process (the smaller of the two)
	numBlocks := min(numBlocksAvailable, numBlocksNeeded)

	// Decode each block
	for i := range numBlocks {
		// Extract 23-bit block
		start := i * 23
		end := start + 23
		block := received[start:end]

		// Calculate remaining data length
		dataStart := i * 12
		decodeLen := 12
		if remaining := len(data) - dataStart; remaining < 12 {
			decodeLen = remaining
		}

		decode(block, data[dataStart:dataStart+decodeLen])
	}

	// If received is insufficient, fill remaining with false
	for i := numBlocks * 12; i < len(data); i++ {
		data[i] = false
	}
}

// encode performs Golay(23,12) encoding.
// Takes 12 bits of data and writes 11 parity bits to parity.
func encode(data []bool, parity []bool) {
	if len(data) != 12 {
		panic("data must be 12 bits")
	}
	if len(parity) != 11 {
		panic("parity must be 11 bits")
	}

	// Calculate parity bits (product of G matrix and data)
	for i := range 11 {
		p := false
		for j := range 12 {
			if data[j] && g[j*11+i] {
				p = !p
			}
		}
		parity[i] = p
	}
}

// decode performs Golay(23,12) decoding (error correction).
// Takes a 23-bit received word, corrects errors, and writes result to data.
// data length must be 12 or less. If less than 12, remaining bits are ignored.
// Can correct up to 3 bit errors.
func decode(received []bool, data []bool) {
	if len(received) != 23 {
		panic("received must be 23 bits")
	}
	if len(data) > 12 {
		panic("data length must be 12 or less")
	}

	// Calculate syndrome S = H^T * r
	syndrome := make([]bool, 11)
	for i := range 11 {
		s := false
		for j := range 23 {
			if received[j] && h[j*11+i] {
				s = !s
			}
		}
		syndrome[i] = s
	}

	// If syndrome is zero, no errors
	isZero := true
	for _, bit := range syndrome {
		if bit {
			isZero = false
			break
		}
	}

	corrected := make([]bool, 23)
	copy(corrected, received)

	if !isZero {
		// Search for error pattern (in order: 1-bit, 2-bit, 3-bit errors)
		corrected = correctErrors(received, syndrome)
	}

	// Write data portion to argument (up to data length)
	copy(data, corrected[:len(data)])
}

// correctErrors performs error correction (detects and corrects 1-3 bit errors).
func correctErrors(received []bool, syndrome []bool) []bool {
	// Check for 1-bit errors
	for pos := range 23 {
		match := true
		for i := range 11 {
			if syndrome[i] != h[pos*11+i] {
				match = false
				break
			}
		}
		if match {
			corrected := make([]bool, 23)
			copy(corrected, received)
			corrected[pos] = !corrected[pos]
			return corrected
		}
	}

	// Check for 2-bit errors
	for i := range 23 {
		for j := i + 1; j < 23; j++ {
			testSyndrome := make([]bool, 11)
			for k := range 11 {
				testSyndrome[k] = xorBool(h[i*11+k], h[j*11+k])
			}

			match := true
			for k := range 11 {
				if testSyndrome[k] != syndrome[k] {
					match = false
					break
				}
			}

			if match {
				corrected := make([]bool, 23)
				copy(corrected, received)
				corrected[i] = !corrected[i]
				corrected[j] = !corrected[j]
				return corrected
			}
		}
	}

	// Check for 3-bit errors
	for i := range 23 {
		for j := i + 1; j < 23; j++ {
			for l := j + 1; l < 23; l++ {
				testSyndrome := make([]bool, 11)
				for k := range 11 {
					testSyndrome[k] = xorBool(xorBool(h[i*11+k], h[j*11+k]), h[l*11+k])
				}

				match := true
				for k := range 11 {
					if testSyndrome[k] != syndrome[k] {
						match = false
						break
					}
				}

				if match {
					corrected := make([]bool, 23)
					copy(corrected, received)
					corrected[i] = !corrected[i]
					corrected[j] = !corrected[j]
					corrected[l] = !corrected[l]
					return corrected
				}
			}
		}
	}

	// If uncorrectable, return original data
	corrected := make([]bool, 23)
	copy(corrected, received)
	return corrected
}

// xorBool is a helper function for XOR operation.
func xorBool(a, b bool) bool {
	return a != b
}
