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

## Implementation

This implementation is based on the generator and parity check matrices from:
- Reference: [zexy-swami/golay_code](https://github.com/zexy-swami/golay_code) (MIT License)

The matrices use the standard Golay(23,12) construction, ensuring all 3-bit error patterns can be corrected.

## License

Apache 2.0 License - see LICENSE file for details.

The reference implementation (zexy-swami/golay_code) is under MIT License, which permits sublicensing under Apache 2.0.

