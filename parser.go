package httpfuzz

import (
	"bufio"
	"net/http"
	"os"
)

// RequestFromFile parses an HTTP request from a file.
func RequestFromFile(filename string) (*http.Request, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(file)
	return http.ReadRequest(reader)
}
