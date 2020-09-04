package httpfuzz

// A DelimiterArray finds the positions of a delimiter within a byte slice.
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
