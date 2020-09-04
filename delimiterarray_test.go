package httpfuzz

import (
	"reflect"
	"testing"
)

func TestDelimiterArrayLookupFindsCorrectOffsets(t *testing.T) {
	contents := []byte("The delimiter is the `backtick` character")
	delimiter := byte('`')
	array := &DelimiterArray{Contents: contents}

	expectedOffsets := []int{21, 30}
	offsets := array.Lookup(delimiter)
	if !reflect.DeepEqual(expectedOffsets, offsets) {
		t.Fatalf("Expected %+v, got %+v", expectedOffsets, offsets)
	}
}
