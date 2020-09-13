package httpfuzz

import (
	"context"
	"fmt"
)

// fuzzerState represents the work to be done by a requestFuzzer at any given time.
type fuzzerState struct {
	Seed                *Request
	PayloadWord         string
	PayloadFile         *File
	BodyTargetDelimiter byte
}

// requestGenerator is a function that takes the state of the fuzzer and sends requests down to the executor based on that, or errors if something went wrong.
// RequestFuzzers should copy the seed request in state before operating on it.
type requestGenerator func(state *fuzzerState, targets []string, jobs chan<- *Job, errors chan<- error)

// fuzzFiles applies a file to every file key we're targeting in the seed request
func fuzzFiles(state *fuzzerState, targets []string, jobs chan *Job, errors chan error) {
	for _, fileKey := range targets {
		req, err := state.Seed.CloneBody(context.Background())
		if err != nil {
			errors <- err
			return
		}

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
