package golay

import "testing"

func TestExhaustive(t *testing.T) {
	var max uint16 = 1<<12 - 1
	for d := range max {
		c := EncodeWord(d)
		// 0bit error
		r := Decode(c)
		if r != d {
			t.Fatalf("Exhaustive decoding failed for data %d: got %d", d, r)
		}
		// 1bit error
		for i := range 23 {
			e := c ^ (1 << i)
			r := Decode(e)
			if r != d {
				t.Fatalf("Exhaustive decoding failed for data %d with 1-bit error at position %d: got %d", d, i, r)
			}
		}
		// 2bit error
		for i := range 23 {
			for j := i + 1; j < 23; j++ {
				e := c ^ (1 << i) ^ (1 << j)
				r := Decode(e)
				if r != d {
					t.Fatalf("Exhaustive decoding failed for data %d with 2-bit errors at positions %d and %d: got %d", d, i, j, r)
				}
			}
		}
		// 3bit error
		for i := range 23 {
			for j := i + 1; j < 23; j++ {
				for k := j + 1; k < 23; k++ {
					e := c ^ (1 << i) ^ (1 << j) ^ (1 << k)
					r := Decode(e)
					if r != d {
						t.Fatalf("Exhaustive decoding failed for data %d with 3-bit errors at positions %d, %d, and %d: got %d", d, i, j, k, r)
					}
				}
			}
		}
	}
}
