# golay
Golay(23,12) encoding and decoding in Go

## Overview

This package provides efficient Golay(23,12) error-correcting code implementation in Go.

- **Encode**: Generate 11-bit parity from 12-bit data
- **EncodeWord**: Generate complete 23-bit codeword
- **Decode**: Decode with up to 3-bit error correction

## Features

- Zero-allocation encoding and decoding
- Exhaustive error correction up to 3 bits (perfect code property)
- Simple and intuitive API

## Usage

```go
import "github.com/yyyoichi/golay"

// Encode 12-bit data
data := uint16(0b101100110011)
parity := golay.Encode(data)           // Returns 11-bit parity
codeword := golay.EncodeWord(data)     // Returns 23-bit codeword

// Decode with error correction
received := codeword ^ 0b111 // Introduce 3-bit error
decoded := golay.Decode(received)      // Returns original data
```

### Stream Processing

For processing binary data streams, this package provides `Encoder` and `Decoder` that work with MSB-aligned data and handle automatic blocking:

```go
// Encode a stream of bytes
data := []uint8{0xFF, 0xF0, 0xAB}
var encoded []uint32
encoder := golay.NewEncoder(&encoded)
err := encoder.Encode(data, 0) // 0 means use all bits
fmt.Println(encoder.Bits())    // Get total encoded bits

// Multiple encodes can append to the same output
moreData := []uint8{0x12, 0x34}
err = encoder.Encode(moreData, 0)

// Decode back to original data
var decoded []uint8
decoder := golay.NewDecoder(encoded, encoder.Bits())
err = decoder.Decode(&decoded)

// Convenience functions are also available
var encoded []uint32
golay.EncodeBinay(data, &encoded)
var decoded []uint8
golay.DecodeBinay(encoded, &decoded)

// Calculate encoded/decoded sizes
inputBits := 100
encodedBits := golay.EncodedBits(inputBits)   // Returns 184 (8 blocks × 23 bits)
decodedBits := golay.DecodedBits(encodedBits) // Returns 96 (8 blocks × 12 bits)
```

The encoder holds a writer internally and can append multiple encode operations to the same output slice. The encoder splits input data into 12-bit blocks and encodes each into a 23-bit codeword. The decoder reverses this process with automatic error correction.

## Implementation

This implementation is based on the generator and parity check matrices from:
- Reference: [zexy-swami/golay_code](https://github.com/zexy-swami/golay_code) (MIT License)

The matrices use the standard Golay(23,12) construction, ensuring all 3-bit error patterns can be corrected.

## License

Apache 2.0 License - see LICENSE file for details.

The reference implementation (zexy-swami/golay_code) is under MIT License, which permits sublicensing under Apache 2.0.

