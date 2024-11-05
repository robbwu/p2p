package encryption

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
)

// GenerateRandomWords: generate 4 random words from
func IntsToWords(indexes []int16) ([]string, error) {
	var words []string
	for _, idx := range indexes {
		if idx < 0 || idx > 2047 {
			return nil, fmt.Errorf("index out of range: %d", idx)
		}
		words = append(words, English[idx])
	}

	return words, nil
}

func WordsToInts([]string) ([]int16, error) {
	var indexes []int16
	for _, word := range English {
		idx, ok := WordToIndex[word]
		if !ok {
			return nil, fmt.Errorf("word not found: %s", word)
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// index: 0-2047; pack into bytes tightly; for example, index only takes 11 bits.
// 4 indexes take 44 bits, which is 6 bytes.
func IntsToBytes(indexes []int16) ([]byte, error) {
	if len(indexes)%4 != 0 {
		return nil, fmt.Errorf("indexes length must be a multiple of 4")
	}

	bytes := make([]byte, len(indexes)/4*6)
	for i := 0; i < len(indexes); i += 4 {
		var bits uint64
		for j := 0; j < 4; j++ {
			bits |= uint64(indexes[i+j]&0x7FF) << (11 * j)
		}
		binary.BigEndian.PutUint64(bytes[i/4*6:], bits)
	}

	return bytes, nil
}

func BytesToInts(bytes []byte) ([]int16, error) {
	if len(bytes)%6 != 0 {
		return nil, fmt.Errorf("bytes length must be a multiple of 6")
	}

	ints := make([]int16, len(bytes)/6*4)
	for i := 0; i < len(bytes); i += 6 {
		bits := binary.BigEndian.Uint64(bytes[i : i+6])
		for j := 0; j < 4; j++ {
			ints[i/6*4+j] = int16(bits & 0x7FF)
			bits >>= 11
		}
	}

	return ints, nil
}

var MaxIndex = big.NewInt(2048)

// generate 4 random words
func GenerateRandomWords() ([]string, error) {
	i1, _ := rand.Int(rand.Reader, MaxIndex)
	i2, _ := rand.Int(rand.Reader, MaxIndex)
	i3, _ := rand.Int(rand.Reader, MaxIndex)
	i4, _ := rand.Int(rand.Reader, MaxIndex)
	var words []string
	words = append(words, English[i1.Int64()])
	words = append(words, English[i2.Int64()])
	words = append(words, English[i3.Int64()])
	words = append(words, English[i4.Int64()])
	return words, nil
}

// WordsToBytes converts 4 BIP39 words into a 6-byte array
// Each word index is 11 bits (0-2047)
// Total bits: 44 (4 * 11), will be stored in 6 bytes (48 bits)
func WordsToBytes(words []string) ([]byte, error) {
	if len(words) != 4 {
		return nil, fmt.Errorf("exactly 4 words required, got %d", len(words))
	}

	// Convert words to their indexes
	indexes := make([]uint16, 4)
	for i, word := range words {
		idx, ok := WordToIndex[word]
		if !ok {
			return nil, fmt.Errorf("word not found: %s", word)
		}
		if idx < 0 || idx >= 2048 {
			return nil, fmt.Errorf("invalid word index: %d", idx)
		}
		indexes[i] = uint16(idx)
	}

	// Allocate 6 bytes for output
	result := make([]byte, 6)

	// Pack the 11-bit indexes into the byte array
	// First word: 11 bits
	result[0] = byte(indexes[0] >> 3)
	result[1] = byte((indexes[0]&0x7)<<5 | (indexes[1] >> 6))
	// Second word: next 11 bits
	result[2] = byte((indexes[1]&0x3F)<<2 | (indexes[2] >> 9))
	result[3] = byte((indexes[2] & 0x1FF) >> 1)
	// Third word: next 11 bits
	result[4] = byte((indexes[2]&0x1)<<7 | (indexes[3] >> 4))
	// Fourth word: final 11 bits
	result[5] = byte((indexes[3] & 0xF) << 4)

	return result, nil
}
