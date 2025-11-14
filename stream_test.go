package golay

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/yyyoichi/bitstream-go"
)

func TestStream(t *testing.T) {
	t.Run("Encode", func(t *testing.T) {
		{
			var v []uint32
			// 16bit -> 2 block -> 23bit x 2 = 46bit -> 2 uint32
			_ = EncodeBinay([]uint8{0xFF, 0xF0}, &v)
			if len(v) != 2 {
				t.Fatalf("EncodeBinay uint8 failed: got length %d, want %d", len(v), 2)
			}
			if v[0] != 0xFFFFFE00 {
				t.Errorf("EncodeBinay uint32 failed: got %#x, want %#x", v[0], 0xFFFFFFE00)
			}
			if v[1] != 0 {
				t.Errorf("EncodeBinay uint32 failed: got %#x, want %#x", v, 0)
			}
		}
		{
			var v []uint32
			// 12bit -> 1 block -> 1 uint32
			enc := NewEncoder([]uint16{0xFFF0}, 12)
			if enc.Bits() != 23 {
				t.Fatalf("Encoder.Bits() failed: got %d, want %d", enc.Bits(), 23)
			}
			_ = enc.Encode(&v)
			if len(v) != 1 {
				t.Fatalf("EncodeBinay uint16 failed: got length %d, want %d", len(v), 1)
			}
			if v[0] != 0xFFFFFE00 {
				t.Errorf("EncodeBinay uint32 failed: got %#x, want %#x", v[0], 0xFFFFFE00)
			}
		}
	})
	t.Run("Decode", func(t *testing.T) {
		{
			var v []uint16
			// 32bit -> 2block -> 24bit -> 2 uint16
			_ = DecodeBinay([]uint32{0xFFFFFE00, 0}, &v)
			if len(v) != 2 {
				t.Fatalf("DecodeBinay uint32 failed: got length %d, want %d", len(v), 2)
			}
			if v[0] != 0xFFF0 {
				t.Errorf("DecodeBinay uint16 failed: got %#x, want %#x", v[0], 0xFFF0)
			}
			if v[1] != 0 {
				t.Errorf("DecodeBinay uint16 failed: got %#x, want %#x", v[1], 0)
			}
		}
		{
			var v []uint16
			// 23bit -> 1block -> 12bit -> 1 uint16
			dec := NewDecoder([]uint32{0xFFFFFE00, 0}, 23)
			if dec.Bits() != 12 {
				t.Fatalf("Decoder.Bits() failed: got %d, want %d", dec.Bits(), 12)
			}
			_ = dec.Decode(&v)
			if len(v) != 1 {
				t.Fatalf("DecodeBinay uint32 failed: got length %d, want %d", len(v), 1)
			}
			if v[0] != 0xFFF0 {
				t.Errorf("DecodeBinay uint16 failed: got %#x, want %#x", v[0], 0xFFF0)
			}
		}
	})
	t.Run("RoundTrip", func(t *testing.T) {
		for range 100 {
			var w = bitstream.NewBitWriter[uint8](0, 0)
			l := rand.Intn(0xFFFF)
			for range l {
				w.Write8(0, 8, uint8(rand.Intn(255)))
			}
			testdata := w.Data()
			bits := w.Bits()
			var encoded []uint8
			enc := NewEncoder(testdata, bits)
			_ = enc.Encode(&encoded)
			encodedBits := enc.Bits()
			var decoded []uint8
			dec := NewDecoder(encoded, encodedBits)
			_ = dec.Decode(&decoded)

			want := bitstream.NewBitReader(testdata, 0, 0)
			got := bitstream.NewBitReader(decoded, 0, 0)
			for i := range l {
				bitWant := want.Read8R(0, 8)
				bitGot := got.Read8R(0, 8)
				if bitWant != bitGot {
					t.Fatalf("RoundTrip failed at index %d: got %#x, want %#x", i, bitGot, bitWant)
				}
			}
			fmt.Print("o")
		}
		fmt.Println()
	})
}
