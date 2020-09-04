package httpfuzz

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
