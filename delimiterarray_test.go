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

func TestDelimiterArrayGetReturnsCorrectOffsets(t *testing.T) {
	contents := []byte("The delimiter is the `backtick` character")
	delimiter := byte('`')
	array := &DelimiterArray{Contents: contents}

	start, end, err := array.Get(0, delimiter)
	if err != nil {
		t.Fatal(err)
	}

	const expectedStart = 21
	const expectedEnd = 30
	if start != expectedStart {
		t.Fatalf("Expected start %d, got %d", expectedStart, start)
	}

	if end != expectedEnd {
		t.Fatalf("Expected end %d got %d", expectedEnd, end)
	}
}
