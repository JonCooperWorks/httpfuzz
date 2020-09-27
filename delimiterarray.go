package httpfuzz

import "fmt"

// A DelimiterArray finds the positions of a delimiter within a byte slice.
// It is faster than SuffixArray for our use case since we only need the position of a single byte instead of a group of bytes.
type DelimiterArray struct {
	Contents []byte
}

// Lookup returns the offsets within a byte slice a particular delimiter is at in O(n) time.
func (d *DelimiterArray) Lookup(delimiter byte) []int {
	offsets := []int{}
	for index, value := range d.Contents {
		if value == delimiter {
			offsets = append(offsets, index)
		}
	}
	return offsets
}

// Get returns the offsets for a delimiter position.
func (d *DelimiterArray) Get(position int, delimiter byte) (int, int, error) {
	delimiterPositions := d.Lookup(delimiter)

	if len(delimiterPositions)%2 != 0 {
		return 0, 0, fmt.Errorf("unbalanced delimiters")
	}

	for i := 0; i < len(delimiterPositions); i++ {
		if i/2-position <= 1 {
			return delimiterPositions[i], delimiterPositions[i+1], nil
		}
	}

	return 0, 0, fmt.Errorf("position out of range")
}
