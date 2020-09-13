package httpfuzz

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
)

const (
	headerLocation         = "header"
	bodyLocation           = "body"
	urlParamLocation       = "url param"
	urlPathArgLocation     = "url path argument"
	directoryRootLocation  = "url directory root"
	directoryRootFieldName = "directory root"
)

// Job represents a request to send with a payload from the fuzzer.
type Job struct {
	Request   *Request
	FieldName string
	Location  string
	Payload   string
}

// Fuzzer creates HTTP requests from a seed request using the combination of inputs specified in the config.
// It uses the producer-consumer pattern efficiently handle large wordlists.
type Fuzzer struct {
	*Config
}

// fuzzerState represents the work to be done by a requestFuzzer at any given time.
type fuzzerState struct {
	Seed                *Request
	PayloadWord         string
	PayloadFile         *File
	BodyTargetDelimiter byte
}

// requestFuzzer is a function that takes the state of the fuzzer and sends requests down to the executor based on that, or errors if something went wrong.
// RequestFuzzers should copy the seed request in state before operating on it.
type requestFuzzer func(state *fuzzerState, targets []string, jobs chan<- *Job, errors chan<- error)

// fuzzFiles applies a file to every file key we're targeting in the seed request
func fuzzFiles(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, fileKey := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		go func(state *fuzzerState, fileKey string, jobs chan<- *Job, errors chan<- error) {
			if state.PayloadFile == nil {
				panic("nil file should never be passed to fuzzFiles")
			}

			file := state.PayloadFile
			err = req.ReplaceMultipartFileData(fileKey, file)
			if err != nil {
				errors <- err
				return
			}

			jobs <- &Job{
				Request:   req,
				FieldName: fileKey,
				Location:  bodyLocation,
				Payload:   file.Name,
			}
		}(state, fileKey, jobs, errors)
	}
}

// fuzzHeaders applies a payload word to every target header in the seed request
func fuzzHeaders(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, header := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		req.Header.Set(header, state.PayloadWord)
		err = req.RemoveDelimiters(state.BodyTargetDelimiter)
		if err != nil {
			errors <- err
			return
		}

		jobs <- &Job{
			Request:   req,
			FieldName: header,
			Location:  headerLocation,
			Payload:   state.PayloadWord,
		}
	}
}

func fuzzURLPathArgs(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, arg := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		req.SetURLPathArgument(arg, state.PayloadWord)
		err = req.RemoveDelimiters(state.BodyTargetDelimiter)
		if err != nil {
			errors <- err
			return
		}

		jobs <- &Job{
			Request:   req,
			FieldName: arg,
			Location:  urlPathArgLocation,
			Payload:   state.PayloadWord,
		}
	}
}

func fuzzDirectoryRoot(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	req, err := state.Seed.CloneBody(context.Background())
	if err != nil {
		errors <- err
		return
	}

	req.SetDirectoryRoot(state.PayloadWord)
	err = req.RemoveDelimiters(state.BodyTargetDelimiter)
	if err != nil {
		errors <- err
		return
	}

	jobs <- &Job{
		Request:   req,
		FieldName: directoryRootFieldName,
		Location:  directoryRootLocation,
		Payload:   state.PayloadWord,
	}
}

func fuzzURLParams(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, param := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		req.SetQueryParam(param, state.PayloadWord)
		err = req.RemoveDelimiters(state.BodyTargetDelimiter)
		if err != nil {
			errors <- err
			return
		}

		jobs <- &Job{
			Request:   req,
			FieldName: param,
			Location:  urlParamLocation,
			Payload:   state.PayloadWord,
		}
	}
}
func fuzzMultipartFormField(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, fieldName := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		err = req.ReplaceMultipartField(fieldName, state.PayloadWord)
		if err != nil {
			errors <- err
			return
		}

		jobs <- &Job{
			Request:   req,
			FieldName: fieldName,
			Location:  bodyLocation,
			Payload:   state.PayloadWord,
		}
	}
}

func fuzzTextBodyWithDelimiters(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	// Fuzz request body injection points
	targetCount, err := state.Seed.BodyTargetCount(state.BodyTargetDelimiter)
	if err != nil {
		errors <- err
		return
	}

	for position := 0; position < targetCount; position++ {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

		err = req.SetBodyPayloadAt(position, state.BodyTargetDelimiter, state.PayloadWord)
		if err != nil {
			errors <- err
			return
		}

		err = req.RemoveDelimiters(state.BodyTargetDelimiter)
		if err != nil {
			errors <- err
			return
		}

		jobs <- &Job{
			Request:   req,
			FieldName: fmt.Sprintf("%d", position),
			Location:  bodyLocation,
			Payload:   state.PayloadWord,
		}
	}
}

// GenerateRequests begins generating HTTP requests based on the seed request and sends them into the returned channel.
// It streams the wordlist from the filesystem line-by-line so it can handle wordlists in constant time.
// The trade-off is that callers cannot know ahead of time how many requests will be sent.
func (f *Fuzzer) GenerateRequests() (<-chan *Job, chan error) {
	jobs := make(chan *Job)
	errors := make(chan error)

	go func(jobs chan *Job, errors chan error) {

		// Send the filesystem stuff independent of the payloads in the wordlist
		for _, filename := range f.FilesystemPayloads {
			file, err := FileFrom(filename, "")
			if err != nil {
				errors <- err
				return
			}

			req, err := f.Seed.CloneBody(context.Background())
			if err != nil {
				errors <- err
				return
			}

			state := &fuzzerState{
				PayloadFile: file,
				Seed:        req,
			}

			fuzzFiles(state, f.TargetFileKeys, jobs, errors)

		}

		if f.EnableGeneratedPayloads {
			for _, fileType := range NativeSupportedFileTypes() {
				req, err := f.Seed.CloneBody(context.Background())
				if err != nil {
					errors <- err
					return
				}

				file, err := GenerateFile(fileType, f.FuzzFileSize, "")
				if err != nil {
					errors <- err
					return
				}

				state := &fuzzerState{
					PayloadFile: file,
					Seed:        req,
				}

				fuzzFiles(state, f.TargetFileKeys, jobs, errors)
			}
		}

		// Generate requests based on the combinations of the headers and URL paths.
		scanner := bufio.NewScanner(f.Wordlist)
		for scanner.Scan() {
			payload := scanner.Text()

			state := &fuzzerState{
				PayloadWord:         payload,
				Seed:                f.Seed,
				BodyTargetDelimiter: f.TargetDelimiter,
			}
			fuzzHeaders(state, f.TargetHeaders, jobs, errors)
			fuzzURLParams(state, f.TargetParams, jobs, errors)
			fuzzURLPathArgs(state, f.TargetPathArgs, jobs, errors)

			if f.FuzzDirectory {
				fuzzDirectoryRoot(state, []string{}, jobs, errors)
			}

			// Prevent delimiter code from firing for multipart requests
			if f.Seed.IsMultipartForm() {
				fuzzMultipartFormField(state, f.TargetMultipartFieldNames, jobs, errors)
			} else {
				fuzzTextBodyWithDelimiters(state, []string{}, jobs, errors)
			}

		}

		// Signal to consumer that we're done
		close(jobs)
		close(errors)

	}(jobs, errors)

	return jobs, errors
}

// RequestCount calculates the total number of requests that will be sent given a set of input and the fields to be fuzzed using combinatorials.
// This will be slower the larger the input file.
// It is imperative that this count matches the number of requests created by GenerateRequest, otherwise httpfuzz will wait forever on requests that aren't coming or exit before all requests are processed.
func (f *Fuzzer) RequestCount() (int, error) {
	count := 1
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

	multipartFieldTargets := len(f.TargetMultipartFieldNames)
	// # of requests = # of lines per file * number of targets
	numRequests := (count * len(f.TargetHeaders)) +
		(count * len(f.TargetParams)) +
		(count * len(f.TargetPathArgs)) +
		multipartFieldTargets*count +
		len(f.FilesystemPayloads)*len(f.TargetFileKeys)

	fileTargets := len(f.TargetFileKeys) * len(NativeSupportedFileTypes())
	if fileTargets > 0 || multipartFieldTargets > 0 {
		if f.EnableGeneratedPayloads {
			numRequests = numRequests + fileTargets
		}
	} else {
		bodyTargetCount, err := f.Seed.BodyTargetCount(f.TargetDelimiter)
		if err != nil {
			return 0, err
		}
		numRequests = numRequests + (count * bodyTargetCount)
	}

	if f.FuzzDirectory {
		numRequests = numRequests + count
	}

	// Move back to the head of the file
	_, err := f.Wordlist.Seek(0, io.SeekStart)
	if err != nil {
		return numRequests, err
	}

	return numRequests, nil
}

// ProcessRequests executes HTTP requests in as they're received over the channel.
func (f *Fuzzer) ProcessRequests(jobs <-chan *Job) {
	for job := range jobs {
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

	f.Logger.Printf("Payload in %s field \"%s\": %s. Received: [%v]", job.Location, job.FieldName, job.Payload, response.StatusCode)

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
