package main

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/joncooperworks/httpfuzz"
)

type bruteForceSuccessful struct {
	logger *log.Logger
}

func (b *bruteForceSuccessful) Listen(results <-chan *httpfuzz.Result) {
	for result := range results {
		// This is a buffer, ReadAll shouldn't fail
		body, _ := ioutil.ReadAll(result.Response.Body)
		if !bytes.Contains(body, []byte("Username and/or password incorrect")) {
			b.logger.Printf("Password found: %s", result.Payload)
		}

	}
}

// New returns a bruteForceSuccessful plugin that detects if a brute force is successful on DWVA.
// This plugin simply logs all output to stdout, but plugins can save requests to disk, database or even send them to other services for further analysis.
func New(logger *log.Logger) (httpfuzz.Listener, error) {
	return &bruteForceSuccessful{logger: logger}, nil
}
