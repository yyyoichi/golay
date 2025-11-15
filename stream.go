package golay

import (
	"errors"
	"reflect"

	"github.com/yyyoichi/bitstream-go"
)

// BinaryValue is a constraint for unsigned integer types.
type BinaryValue interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~uint
}

// EncodeBinay performs Golay encoding on MSB-aligned data by splitting it into 12-bit blocks
// and stores the result in v. Each 12-bit block is encoded into a 23-bit Golay codeword
// (12 data bits + 11 parity bits).
// The input data type I and output type O can be different BinaryValue types.
func EncodeBinay[I, O BinaryValue](data []I, v *[]O) error {
	encoder := NewEncoder(v)
	return encoder.Encode(data, 0)
}

// Encoder performs Golay encoding on MSB-aligned binary data.
// It writes encoded data to an output slice specified at creation time.
// Input data is split into 12-bit blocks, and each block is encoded
// into a 23-bit Golay codeword (12 data bits + 11 parity bits).
// Multiple Encode calls can be made to append additional encoded data.
type Encoder[T BinaryValue] struct {
	writer interface {
		Write16(int, int, uint16)
		AnyData() any
	}
	outputPtr *[]T
	bits      int
}

// NewEncoder creates a new Encoder that writes encoded data to v.
// v must be a pointer to a slice of BinaryValue type where the encoded result will be stored.
// The output type T can be flexibly specified (e.g., *[]uint32, *[]uint8).
// Multiple Encode calls can be made on the same Encoder to append encoded data.
func NewEncoder[T BinaryValue](v *[]T) *Encoder[T] {
	if v == nil {
		panic("v must not be nil")
	}
	var writer interface {
		Write16(int, int, uint16)
		AnyData() any
	}
	var zero T
	switch any(zero).(type) {
	case uint64:
		writer = bitstream.NewBitWriter[uint64](0, 0)
	case uint32:
		writer = bitstream.NewBitWriter[uint32](0, 0)
	case uint16:
		writer = bitstream.NewBitWriter[uint16](0, 0)
	case uint8:
		writer = bitstream.NewBitWriter[uint8](0, 0)
	case uint:
		writer = bitstream.NewBitWriter[uint](0, 0)
	default:
		panic("slice element type must satisfy BinaryValue constraint")
	}
	return &Encoder[T]{
		writer:    writer,
		outputPtr: v,
	}
}

// Encode performs Golay encoding on the given data and appends the result to the output slice.
// data must be a slice of BinaryValue type ([]uint8, []uint16, []uint32, []uint64, or []uint).
// The bits parameter specifies how many bits in the input data are valid.
// For example, if data contains 64-bit values but only 12 bits are valid,
// setting bits=12 results in only one Golay encoding operation instead of six.
// If bits is 0, all bits in the data are considered valid.
// This method can be called multiple times to encode different data into the same output slice.
func (e *Encoder[T]) Encode(data any, bits int) error {
	if data == nil {
		return errors.New("data must not be nil")
	}

	// Create reader based on input data type
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return errors.New("data must be a slice")
	}

	var reader interface {
		SetBits(int)
		Read16R(int, int) uint16
		Bits() int
	}

	switch rv.Type().Elem().Kind() {
	case reflect.Uint64:
		d := rv.Interface().([]uint64)
		reader = bitstream.NewBitReader(d, 0, 0)
	case reflect.Uint32:
		d := rv.Interface().([]uint32)
		reader = bitstream.NewBitReader(d, 0, 0)
	case reflect.Uint16:
		d := rv.Interface().([]uint16)
		reader = bitstream.NewBitReader(d, 0, 0)
	case reflect.Uint8:
		d := rv.Interface().([]uint8)
		reader = bitstream.NewBitReader(d, 0, 0)
	case reflect.Uint:
		d := rv.Interface().([]uint)
		reader = bitstream.NewBitReader(d, 0, 0)
	default:
		return errors.New("data slice element type must satisfy BinaryValue constraint")
	}
	if bits > 0 {
		reader.SetBits(bits)
	}

	numBlocks := (reader.Bits() + 11) / 12
	for i := range numBlocks {
		b := reader.Read16R(12, i)
		// right 12 bits are data
		e.writer.Write16(4, 12, b)
		p := Encode(b)
		// right 11 bits are parity
		e.writer.Write16(5, 11, p)
	}

	e.bits += numBlocks * 23

	// Write result back to the output slice
	result := e.writer.AnyData()
	resultRV := reflect.ValueOf(result)
	outputRV := reflect.ValueOf(e.outputPtr).Elem()
	outputRV.Set(resultRV)

	return nil
}

// Bits returns the total number of bits that have been encoded so far.
// This accumulates across multiple Encode calls on the same Encoder.
// Each 12-bit input block is encoded into a 23-bit Golay codeword.
// For example, if 13 bits of input data are encoded, it results in 2 blocks (23 bits × 2 = 46 bits).
func (e *Encoder[T]) Bits() int {
	return e.bits
}

// EncodedBits calculates the number of bits that would result from encoding
// the given number of input bits using Golay encoding.
// Each 12-bit input block is encoded into a 23-bit Golay codeword.
// The calculation rounds up to encode as many complete blocks as possible.
// For example, 13 bits of input data will be encoded as 2 blocks (23 bits × 2 = 46 bits).
func EncodedBits(bits int) int {
	return (bits + 11) / 12 * 23
}

// DecodeBinay performs Golay decoding on MSB-aligned data by splitting it into 23-bit blocks
// and stores the result in v. Each 23-bit Golay codeword is decoded into a 12-bit data block.
func DecodeBinay[I, O BinaryValue](data []I, v *[]O) error {
	decoder := NewDecoder(data, 0)
	return decoder.Decode(v)
}

// Decoder performs Golay decoding on MSB-aligned binary data.
// It splits the input data into 23-bit blocks and decodes each block
// into a 12-bit data value.
type Decoder[T BinaryValue] struct {
	reader *bitstream.BitReader[T]
}

// NewDecoder creates a new Decoder for MSB-aligned data.
// The bits parameter specifies how many bits in the input data are valid.
// It expects bits to be a multiple of 23; any remainder will be ignored.
// For example, if data contains 64-bit values but only 23 bits are valid,
// setting bits=23 results in only one Golay decoding operation instead of two.
func NewDecoder[T BinaryValue](data []T, bits int) *Decoder[T] {
	reader := bitstream.NewBitReader(data, 0, 0)
	if bits > 0 {
		reader.SetBits(bits)
	}
	return &Decoder[T]{
		reader: reader,
	}
}

// Decode performs Golay decoding and stores the result in v.
// v must be a pointer to a slice of BinaryValue type.
// The output type can be flexibly specified (e.g., *[]uint32, *[]uint8).
func (d *Decoder[T]) Decode(v any) error {
	if v == nil {
		return errors.New("v must not be nil")
	}
	// Type check: ensure v is a pointer
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("v must be a pointer to a slice")
	}
	// Ensure the pointer points to a slice
	elem := rv.Elem()
	if elem.Kind() != reflect.Slice {
		return errors.New("v must be a pointer to a slice")
	}
	var writer interface {
		Write16(int, int, uint16)
		AnyData() any
	}
	elemType := elem.Type().Elem()
	switch elemType.Kind() {
	case reflect.Uint64:
		writer = bitstream.NewBitWriter[uint64](0, 0)
	case reflect.Uint32:
		writer = bitstream.NewBitWriter[uint32](0, 0)
	case reflect.Uint16:
		writer = bitstream.NewBitWriter[uint16](0, 0)
	case reflect.Uint8:
		writer = bitstream.NewBitWriter[uint8](0, 0)
	case reflect.Uint:
		writer = bitstream.NewBitWriter[uint](0, 0)
	default:
		// Ensure the slice element type satisfies BinaryValue constraint
		return errors.New("slice element type must satisfy BinaryValue constraint")
	}

	numBlocks := d.reader.Bits() / 23
	for i := range numBlocks {
		cw := d.reader.Read32R(23, i)
		b := Decode(cw)
		// right 12 bits are data
		writer.Write16(4, 12, b)
	}
	data := writer.AnyData()
	rv.Elem().Set(reflect.ValueOf(data))
	return nil
}

// Bits returns the total number of bits in the decoded output.
// This method decodes only complete 23-bit blocks, discarding any incomplete data.
// For example, 48 bits of input data will be decoded as 2 blocks (12 bits × 2 = 24 bits),
// and the remaining 2 bits will be ignored.
func (d *Decoder[T]) Bits() int {
	return d.reader.Bits() / 23 * 12
}

// DecodedBits calculates the number of bits that would result from decoding
// the given number of encoded bits using Golay decoding.
// Each 23-bit Golay codeword is decoded into a 12-bit data block.
// Only complete 23-bit blocks are decoded; any remainder is discarded.
// For example, 48 bits of encoded data will be decoded as 2 blocks (12 bits × 2 = 24 bits),
// and the remaining 2 bits will be ignored.
func DecodedBits(bits int) int {
	return bits / 23 * 12
}
