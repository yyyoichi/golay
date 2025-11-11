package golay

import (
	"testing"
)

func TestEncoder(t *testing.T) {
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
		_ = NewEncoder([]uint16{0xFFF0}, 12).Encode(&v)
		if len(v) != 1 {
			t.Fatalf("EncodeBinay uint16 failed: got length %d, want %d", len(v), 1)
		}
		if v[0] != 0xFFFFFE00 {
			t.Errorf("EncodeBinay uint32 failed: got %#x, want %#x", v[0], 0xFFFFFE00)
		}
	}
}
