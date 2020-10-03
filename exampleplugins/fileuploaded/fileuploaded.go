package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/joncooperworks/httpfuzz"
)

type fileUploaded struct {
	logger *log.Logger
}

func (b *fileUploaded) Listen(results <-chan *httpfuzz.Result) {
	for result := range results {
		// This is a buffer, ReadAll shouldn't fail
		body, _ := ioutil.ReadAll(result.Response.Body)
		if bytes.Contains(body, []byte("successfully uploaded!")) {
			b.logger.Printf("%s successfully uploaded", result.Payload)
		}

	}
}

// New returns a fileUploaded plugin that detects if a file payload has been uploaded successfully to DVWA.
func New(logger *log.Logger) (httpfuzz.Listener, error) {
	return &fileUploaded{logger: logger}, nil
}
