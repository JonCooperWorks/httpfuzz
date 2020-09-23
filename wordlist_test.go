package httpfuzz

import (
	"os"
	"testing"
)

func TestWordlistCountIsAccurate(t *testing.T) {
	wlFile, err := os.Open("./testdata/useragents.txt")
	if err != nil {
		t.Fatal(err)
	}

	wordlist := &Wordlist{File: wlFile}
	count, err := wordlist.Count()
	if err != nil {
		t.Fatal(err)
	}

	const expectedCount = 5
	if count != expectedCount {
		t.Fatalf("Expected %d, got %d", expectedCount, count)
	}

	wordsReceived := 0
	for range wordlist.Stream() {
		wordsReceived++
		if wordsReceived > count {
			t.Fatalf("Expected %d words, got %d", count, wordsReceived)
		}
	}

	if wordsReceived < count {
		t.Fatalf("Expected %d words, got %d", count, wordsReceived)
	}
}
