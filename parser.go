package httpfuzz

import (
	"bufio"
	"net/http"
	"os"
)

// RequestFromFile parses an HTTP request from a file.
func RequestFromFile(filename string) (*Request, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	req, err := http.ReadRequest(bufio.NewReader(file))
	if err != nil {
		return nil, err
	}

	return &Request{req}, nil
}
