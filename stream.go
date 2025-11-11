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
func EncodeBinay[I, O BinaryValue](data []I, v *[]O) error {
	encoder := NewEncoder(data, 0)
	return encoder.Encode(v)
}

// Encoder performs Golay encoding on MSB-aligned binary data.
// It splits the input data into 12-bit blocks and encodes each block
// into a 23-bit Golay codeword.
type Encoder[T BinaryValue] struct {
	reader *bitstream.BitReader[T]
}

// NewEncoder creates a new Encoder for MSB-aligned data.
// The bits parameter specifies how many bits in the input data are valid.
// For example, if data contains 64-bit values but only 12 bits are valid,
// setting bits=12 results in only one Golay encoding operation instead of six.
func NewEncoder[T BinaryValue](data []T, bits int) *Encoder[T] {
	reader := bitstream.NewBitReader(data, 0, 0)
	if bits > 0 {
		reader.SetBits(bits)
	}
	return &Encoder[T]{
		reader: reader,
	}
}

// Encode performs Golay encoding and stores the result in v.
// v must be a pointer to a slice of BinaryValue type.
// The output type can be flexibly specified (e.g., *[]uint32, *[]uint8).
func (e *Encoder[T]) Encode(v any) error {
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
		U16(int, int, uint16)
		AnyData() (any, int)
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

	numBlocks := (e.reader.Bits() + 11) / 12
	for i := range numBlocks {
		b := e.reader.U16R(12, i)
		// right 12 bits are data
		writer.U16(4, 12, b)
		p := Encode(b)
		// right 11 bits are parity
		writer.U16(5, 11, p)
	}

	data, _ := writer.AnyData()
	rv.Elem().Set(reflect.ValueOf(data))
	return nil
}
