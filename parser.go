package httpfuzz

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

	// Since we're letting the user specify injection points with a delimiter, the content length in the header will not match the body.
	// Making them fix it by hand is awful so let's calculate it for them.
	// This code only runs once at program startup: we're taking the performance hit now so the rest of the program can be efficient and usable.
	diskRequestBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Move back to the head of the file so Go's parser can read it
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	req, err := http.ReadRequest(bufio.NewReader(file))
	if err != nil {
		return nil, err
	}

	// We don't need to hack up the request body if there isn't one.
	if req.ContentLength == 0 {
		// Wrap the request in our native Request type for the rest of the program.
		return &Request{req}, nil
	}

	// Replace the body in the seed request with the body on disk and adjust the content length.
	// I know this is hacky, but there's tests and this is easier than reimplimenting http.ReadRequest.
	bodyOffset := bytes.Index(diskRequestBytes, []byte("\n\n"))
	if bodyOffset == -1 {
		// Check for unix line endings too
		bodyOffset = bytes.Index(diskRequestBytes, []byte("\r\n\r\n"))
		if bodyOffset == -1 {
			return nil, fmt.Errorf("invalid HTTP request provided")
		}
	}

	diskBodyBytes := diskRequestBytes[bodyOffset:len(diskRequestBytes)]
	req.Body = ioutil.NopCloser(bytes.NewReader(diskBodyBytes))
	req.ContentLength = int64(len(diskBodyBytes))

	// Wrap the request in our native Request type for the rest of the program.
	return &Request{req}, nil
}
