package golay

import (
	"slices"
)

// Encoder encodes data into Golay(23,12) codewords.
// Supports writing data in various formats and retrieves encoded codewords.
type Encoder struct {
	blocks []uint32 // Encoded 23-bit codewords
}

// NewEncoder creates a new Encoder instance.
func NewEncoder() *Encoder {
	return &Encoder{
		blocks: make([]uint32, 0),
	}
}

// WriteBytes encodes data from a byte slice.
// bits specifies the total number of data bits to encode from the byte slice.
// Data is split into 12-bit chunks and encoded into 23-bit codewords.
// If bits is not a multiple of 12, the last chunk is padded with zeros.
//
// Example:
//
//	enc.WriteBytes(48, []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC})
//	// 48 bits → 4 blocks (48 bits exactly)
func (e *Encoder) WriteBytes(bits int, data []byte) error {
	n := (bits + 11) / 12
	e.blocks = slices.Grow(e.blocks, n)
	for i := range n {
		start := min(i*12, bits)
		end := min(start+12, bits)
		var block uint16 = 0
		for j := start; j < end; j++ {
			block <<= 1
			if data[j/8]&byte(1<<(7-(j%8))) != 0 {
				block |= 1
			}
		}
		e.blocks = append(e.blocks, EncodeWord(block))
	}
	return nil
}

// WriteU64s encodes data from uint64 values.
// bits specifies the total number of data bits across all values.
// Data is split into 12-bit chunks and encoded into 23-bit codewords.
// If bits is not a multiple of 12, the last chunk is padded with zeros.
//
// Example:
//
//	enc.WriteU64s(112, 0x123456789ABCDEF0, 0xFEDCBA9876543210)
//	// 112 bits → 10 blocks (120 bits with 8-bit padding)
func (e *Encoder) WriteU64s(bits int, data ...uint64) error {
	n := (bits + 11) / 12
	e.blocks = slices.Grow(e.blocks, n)
	for i := range n {
		start := min(i*12, bits)
		end := min(start+12, bits)
		var block uint16 = 0
		for j := start; j < end; j++ {
			block <<= 1
			if (data[j/64] & (1 << (63 - (j % 64)))) != 0 {
				block |= 1
			}
		}
		e.blocks = append(e.blocks, EncodeWord(block))
	}
	return nil
}

// WriteBools encodes data from boolean values.
// Each boolean represents one bit of data.
// Data is split into 12-bit chunks and encoded into 23-bit codewords.
// If the length is not a multiple of 12, the last chunk is padded with false.
//
// Example:
//
//	enc.WriteBools(true, false, true, true, false, ...)
func (e *Encoder) WriteBools(data ...bool) error {
	bits := len(data)
	n := (bits + 11) / 12
	e.blocks = slices.Grow(e.blocks, n)
	for i := range n {
		start := i * 12
		end := min(start+12, bits)
		var block uint16 = 0
		for j := start; j < end; j++ {
			block <<= 1
			if data[j] {
				block |= 1
			}
		}
		e.blocks = append(e.blocks, EncodeWord(block))
	}
	return nil
}

// Codewords returns all encoded 23-bit codewords.
// Each element represents one 23-bit codeword (in the lower 23 bits of uint32).
func (e *Encoder) Codewords() []uint32 {
	return e.blocks
}

// Bytes returns the encoded codewords as a tightly packed byte slice.
// Codewords are packed from MSB (left-aligned) without gaps between codewords.
// Bits flow continuously across byte boundaries.
//
// For example, with 2 codewords (46 bits total):
//
//	Codeword1: [22...0] (23 bits), Codeword2: [22...0] (23 bits)
//	Packed:    [C1:22-15][C1:14-7][C1:6-0,C2:22-20][C2:19-12][C2:11-4][C2:3-0,pad:0000]
//	Result:    [byte0  ][byte1  ][byte2          ][byte3   ][byte4  ][byte5        ]
//
// Total bytes = ceil(total_bits / 8)
func (e *Encoder) Bytes() []byte {
	bits := e.Bits()
	total := (bits + 7) / 8
	result := make([]byte, total)
	for i, cw := range e.blocks {
		bitPos := i * 23
		for j := range 23 {
			if (cw & (1 << (22 - j))) != 0 {
				result[(bitPos+j)/8] |= byte(1 << ((bitPos + j) % 8))
			}
		}
	}
	return nil
}

// Bools returns the encoded codewords as a boolean slice.
// Each boolean represents one bit, packed continuously without padding.
// The length of the returned slice is (number of codewords × 23).
func (e *Encoder) Bools() []bool {
	bits := e.Bits()
	result := make([]bool, bits)
	for i, cw := range e.blocks {
		bitPos := i * 23
		for j := range 23 {
			if (cw & (1 << (22 - j))) != 0 {
				result[bitPos+j] = true
			}
		}
	}
	return nil
}

// Uint64s returns the encoded codewords as a uint64 slice.
// Codewords are packed from MSB (left-aligned) continuously across uint64 boundaries.
// Similar to Bytes(), but uses 64-bit words instead of 8-bit bytes.
//
// For example, with 3 codewords (69 bits total):
//
//	Result: [uint64_0: 64 bits from MSB][uint64_1: 5 bits from MSB + 59 zero padding]
//
// Total elements = ceil(total_bits / 64)
func (e *Encoder) Uint64s() []uint64 {
	bits := e.Bits()
	total := (bits + 63) / 64
	result := make([]uint64, total)
	for i, cw := range e.blocks {
		bitPos := i * 23
		for j := range 23 {
			if (cw & (1 << (22 - j))) != 0 {
				result[(bitPos+j)/64] |= uint64(1 << ((bitPos + j) % 64))
			}
		}
	}
	return result
}

// Bits returns the total number of encoded bits.
// This is the number of codewords multiplied by 23.
func (e *Encoder) Bits() int {
	return len(e.blocks) * 23
}

// Reset clears all encoded data.
func (e *Encoder) Reset() {
	e.blocks = e.blocks[:0]
}

// Decoder decodes Golay(23,12) codewords with error correction.
// Receives encoded data via Write methods, decodes it with error correction,
// and provides decoded data in various formats.
type Decoder struct {
	decoded []uint16 // Decoded 12-bit data blocks
}

// NewDecoder creates a new Decoder instance.
func NewDecoder() *Decoder {
	return &Decoder{
		decoded: make([]uint16, 0),
	}
}

// WriteCodewords decodes 23-bit codewords with error correction and stores the result.
// Each codeword is decoded into 12-bit data.
func (d *Decoder) WriteCodewords(codewords ...uint32) error {
	for _, cw := range codewords {
		d.decoded = append(d.decoded, Decode(cw))
	}
	return nil
}

// WriteBytes decodes data from a byte slice.
// bits specifies the total number of bits in the encoded data (must be multiple of 23).
// Expects bytes to be MSB-aligned (left-aligned) as produced by Encoder.Bytes().
// Codewords are extracted continuously from the MSB, decoded, and stored.
//
// Example:
//
//	dec.WriteBytes(184, encodedBytes)  // 184 bits = 8 codewords
func (d *Decoder) WriteBytes(bits int, data []byte) error {
	l := len(data) * 8
	n := bits / 23
	for i := range n {
		var cw uint32 = 0
		start := min(i*23, l)
		end := min(start+23, l)
		for j := start; j < end; j++ {
			cw <<= 1
			if (data[j/8] & byte(1<<7-(j%8))) != 0 {
				cw |= 1
			}
		}
		d.decoded = append(d.decoded, Decode(cw))
	}
	return nil
}

// WriteUint64s decodes data from uint64 values.
// bits specifies the total number of bits in the encoded data (must be multiple of 23).
// Expects uint64s to be MSB-aligned (left-aligned) as produced by Encoder.Uint64s().
// Codewords are extracted continuously from the MSB, decoded, and stored.
//
// Example:
//
//	dec.WriteUint64s(184, data...)  // 184 bits = 8 codewords
func (d *Decoder) WriteUint64s(bits int, data ...uint64) error {
	l := len(data) * 64
	n := bits / 23
	for i := range n {
		var cw uint32 = 0
		start := min(i*23, l)
		end := min(start+23, l)
		for j := start; j < end; j++ {
			cw <<= 1
			if (data[j/64] & (1 << (63 - (j % 64)))) != 0 {
				cw |= 1
			}
		}
		d.decoded = append(d.decoded, Decode(cw))
	}
	return nil
}

// WriteBools decodes data from boolean values.
// The length of data must be a multiple of 23 (each 23 bools = 1 codeword).
// Each 23 consecutive booleans form one codeword (MSB first).
// Codewords are decoded and stored.
//
// Example:
//
//	dec.WriteBools(bools...)  // len(bools) must be multiple of 23
func (d *Decoder) WriteBools(data ...bool) error {
	bits := len(data)
	n := bits / 23
	for i := range n {
		var cw uint32 = 0
		start := i * 23
		end := start + 23
		for j := start; j < end; j++ {
			cw <<= 1
			if data[j] {
				cw |= 1
			}
		}
		d.decoded = append(d.decoded, Decode(cw))
	}
	return nil
}

// Bytes returns all decoded data as a byte slice.
// Data is packed from MSB (left-aligned) continuously across byte boundaries.
// Returns the complete decoded data.
//
// Example:
//
//	data, err := dec.Bytes()
//	// Returns all decoded bits packed into bytes
func (d *Decoder) Bytes() ([]byte, error) {
	bits := d.Bits()
	total := (bits + 7) / 8
	result := make([]byte, total)
	for i, block := range d.decoded {
		bitPos := i * 12
		for j := range 12 {
			if (block & (1 << (11 - j))) != 0 {
				result[(bitPos+j)/8] |= byte(1 << ((bitPos + j) % 8))
			}
		}
	}
	return result, nil
}

// Uint64s returns all decoded data as a uint64 slice.
// Data is packed from MSB (left-aligned) continuously across uint64 boundaries.
// Returns the complete decoded data.
//
// Example:
//
//	data, err := dec.Uint64s()
//	// Returns all decoded bits packed into uint64 values
func (d *Decoder) Uint64s() ([]uint64, error) {
	bits := d.Bits()
	total := (bits + 63) / 64
	result := make([]uint64, total)
	for i, block := range d.decoded {
		bitPos := i * 12
		for j := range 12 {
			if (block & (1 << (11 - j))) != 0 {
				result[(bitPos+j)/64] |= uint64(1 << ((bitPos + j) % 64))
			}
		}
	}
	return result, nil
}

// Bools returns all decoded data as a boolean slice.
// Each boolean represents one decoded bit.
// Returns the complete decoded data.
//
// Example:
//
//	data, err := dec.Bools()
//	// Returns all decoded bits as boolean values
func (d *Decoder) Bools() ([]bool, error) {
	bits := d.Bits()
	result := make([]bool, bits)
	for i, block := range d.decoded {
		bitPos := i * 12
		for j := range 12 {
			if (block & (1 << (11 - j))) != 0 {
				result[bitPos+j] = true
			}
		}
	}
	return result, nil
}

// Bits returns the total number of decoded bits.
// This is the number of decoded blocks multiplied by 12.
func (d *Decoder) Bits() int {
	return len(d.decoded) * 12
}

// Reset clears all decoded data.
func (d *Decoder) Reset() {
	d.decoded = d.decoded[:0]
}
