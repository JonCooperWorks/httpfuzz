package httpfuzz

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"time"
)

const (
	headerLocation   = "header"
	bodyLocation     = "body"
	urlParamLocation = "urlParam"
)

// Job represents a request to send with a payload from the fuzzer.
type Job struct {
	Request   *Request
	FieldName string
	Location  string
	Payload   string
}

// Fuzzer creates HTTP requests from a seed request using the combination of inputs specified in the config.
// It uses the producer-consumer pattern efficiently handle large inputs.
type Fuzzer struct {
	*Config
}

// GenerateRequests begins generating HTTP requests based on the seed request and sends them into the returned channel.
// It streams the wordlist from the filesystem line-by-line so it can handle wordlists in constant time.
// The trade-off is that callers cannot know ahead of time how many requests will be sent.
func (f *Fuzzer) GenerateRequests() <-chan *Job {
	requestQueue := make(chan *Job)

	go func(requestQueue chan *Job) {
		// Generate requests based on the combinations of the headers and URL paths.
		scanner := bufio.NewScanner(f.Wordlist)
		for scanner.Scan() {
			payload := scanner.Text()

			// Send requests with each of the headers in the request.
			for _, header := range f.TargetHeaders {
				req, err := f.Seed.CloneBody(context.Background())
				if err != nil {
					continue
				}

				req.Header.Set(header, payload)

				// TODO: Send payload and header down chan too
				// Push generated requests into the queue as they are created
				requestQueue <- &Job{
					Request:   req,
					FieldName: header,
					Location:  headerLocation,
					Payload:   payload,
				}
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
	count = (count + 1) * len(f.TargetHeaders)

	// Move back to the head of the file
	_, err := f.Wordlist.Seek(0, io.SeekStart)
	if err != nil {
		return count, err
	}

	return count, nil
}

// ProcessRequests executes HTTP requests in as they're received over the channel.
func (f *Fuzzer) ProcessRequests(requestQueue <-chan *Job) {
	for job := range requestQueue {
		if job == nil {
			// A nil job signals that the producer is finished.
			break
		}

		go f.requestWorker(job)

		// If there's no delay, it'll return immediately, so we don't need to waste time checking.
		time.Sleep(f.RequestDelay)
	}

	f.waitGroup.Wait()
}

func (f *Fuzzer) requestWorker(job *Job) {
	defer f.waitGroup.Done()

	job.Request.URL.Scheme = f.URLScheme

	// Keep the request body around for the plugins.
	request, err := job.Request.CloneBody(context.Background())
	if err != nil {
		f.Logger.Printf("Error cloning request body: %v", err)
		return
	}

	response, err := f.Client.Do(job.Request)
	if err != nil {
		f.Logger.Printf("Error sending request: %v", err)
		return
	}

	f.Logger.Printf("Payload in %s: %s. Received: [%v]", job.Location, job.Payload, response.StatusCode)

	for _, plugin := range f.Plugins {
		req, err := request.CloneBody(context.Background())
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
		result := &Result{
			Request:   req,
			Response:  resp,
			Payload:   job.Payload,
			Location:  job.Location,
			FieldName: job.FieldName,
		}
		go f.runPlugin(plugin, result)
	}
}

// WaitFor adds the requests the fuzzer will send to our internal sync.WaitGroup.
// This keeps the fuzzer running until all requests have been completed.
func (f *Fuzzer) WaitFor(requests int) {
	f.waitGroup.Add(requests)
}

func (f *Fuzzer) runPlugin(plugin Plugin, result *Result) {
	err := plugin.OnSuccess(result)
	if err != nil {
		f.Logger.Printf("Error running plugin %s: %v", plugin.Name(), err)
	}
}
