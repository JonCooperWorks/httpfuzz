package httpfuzz

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

// Fuzzer creates HTTP requests from a seed request using the combination of inputs specified in the config.
// It uses the producer-consumer pattern efficiently handle large inputs.
type Fuzzer struct {
	*Config
}

// GenerateRequests begins generating HTTP requests based on the seed request and sends them into the returned channel.
// It streams the wordlist from the filesystem line-by-line so it can handle wordlists in constant time.
// The trade-off is that callers cannot know ahead of time how many requests will be sent.
func (f *Fuzzer) GenerateRequests() <-chan *Request {
	requestQueue := make(chan *Request)

	go func(requestQueue chan *Request) {
		// Generate requests based on the combinations of the headers and URL paths.
		scanner := bufio.NewScanner(f.Wordlist)
		for scanner.Scan() {
			word := scanner.Text()

			// Send requests with each of the headers in the request.
			for _, header := range f.TargetHeaders {
				req, err := f.Seed.CloneBody(context.Background())
				if err != nil {
					continue
				}

				req.Header.Set(header, word)

				// Push generated requests into the queue as they are created
				requestQueue <- req
			}
		}

		// Signal to consumer that we're done
		requestQueue <- nil

	}(requestQueue)

	return requestQueue
}

// RequestCount calculates the total number of requests that will be sent given a set of input and the fields to be fuzzed using combinatorials.
// This will be slower the larger the input file.
func (f *Fuzzer) RequestCount() (int, error) {
	var count int
	const lineBreak = '\n'

	buf := make([]byte, bufio.MaxScanTokenSize)

	for {
		bufferSize, err := f.Wordlist.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], lineBreak)
			if i == -1 || bufferSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
		if err == io.EOF {
			break
		}
	}

	// # of requests = # of lines per file * number of targets
	count = count * len(f.TargetHeaders)
	return count, nil
}

// ProcessRequests executes HTTP requests in the background as they're received over the channel.
func (f *Fuzzer) ProcessRequests(requestQueue <-chan *Request) {
	for req := range requestQueue {
		if req == nil {
			// A nil request signals that this is the last request.
			break
		}

		// Handle each request in the background and move to the next one.
		if f.MaxConcurrentRequests == 0 {
			go f.requestWorker(req)
			continue
		}

		// TODO: Use a goroutine pool if the user set the max # of concurrent requests
	}
}

func (f *Fuzzer) requestWorker(request *Request) {
	// Keep the request body around for the plugins.
	req, err := request.CloneBody(context.Background())
	if err != nil {
		f.Logger.Printf("Error cloning request body: %v", err)
		return
	}

	response, err := f.Client.Do(request)
	if err != nil {
		return
	}

	for _, plugin := range f.Plugins {
		r, err := req.CloneBody(context.Background())
		if err != nil {
			f.Logger.Printf("Error cloning request for plugin %s: %v", plugin.Name(), err)
			continue
		}

		resp, err := response.CloneBody()
		if err != nil {
			f.Logger.Printf("Error cloning response for plugin %s: %v", plugin.Name(), err)
			continue
		}

		// Run each plugin in its own goroutine
		go plugin.OnSuccess(r, resp)
	}
}
