package httpfuzz

import (
	"net/http"
)

// Fuzzer creates HTTP requests from a seed request using the combination of inputs specified in the config.
// It uses the producer-consumer pattern efficiently handle large inputs.
type Fuzzer struct {
	*Config
}

// GenerateRequests begins generating HTTP requests based on the seed request and sends them into the returned channel.
// It streams all files from the filesystem line-by-line so it can handle wordlists in constant time.
// The trade-off is that callers cannot know ahead of time how many requests will be sent.
func (f *Fuzzer) GenerateRequests() chan<- *http.Request {
	requestQueue := make(chan *http.Request)

	// TODO: Generate requests based on the combinations of the headers and URL paths.
	// TODO: Push generated requests into the queue as they are created

	return requestQueue
}

// RequestCount calculates the total number of requests that will be sent given a set of input and the fields to be fuzzed using combinatorials.
// This will be slower the larger the input files.
func (f *Fuzzer) RequestCount() (int, error) {
	return 0, nil
}

// ProcessRequests executes HTTP requests in the background as they're received over the channel.
func (f *Fuzzer) ProcessRequests(requestQueue <-chan *http.Request) {
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

func (f *Fuzzer) requestWorker(req *http.Request) {
	resp, err := f.Client.Do(req)
	if err != nil {
		for _, plugin := range f.Plugins {
			plugin.OnError(req, err)
		}
		return
	}

	for _, plugin := range f.Plugins {
		plugin.OnSuccess(req, resp)
	}
}
