package httpfuzz

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

// Client is a modified net/http Client that can natively handle our request and response types
type Client struct {
	*http.Client
}

// Do wraps Go's net/http client with our Request and Response types.
func (c *Client) Do(req *Request) (*Response, error) {
	resp, err := c.Client.Do(req.Request)
	return &Response{Response: resp}, err
}

// Request is a more fuzzable *http.Request.
// It supports deep-cloning its body and has several convenience methods for modifying request attributes.
type Request struct {
	*http.Request
}

// CloneBody makes a copy of a request, including its body, while leaving the original body intact.
func (r *Request) CloneBody(ctx context.Context) (*Request, error) {
	req := &Request{Request: r.Request.Clone(ctx)}

	// We have to manually set the host in the URL.
	req.URL.Host = r.Request.Host

	// Prevent an error when sending the request
	req.RequestURI = ""
	if req.Body == nil {
		return req, nil
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return req, err
	}

	// Put back the original body
	r.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Clone the request body
	req.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
	return req, nil
}

// HasPathArgument returns true if a request URL has a given path argument.
func (r *Request) HasPathArgument(pathArg string) bool {
	path := strings.Split(r.URL.EscapedPath(), "/")
	for _, arg := range path {
		if arg == pathArg {
			return true
		}
	}

	return false
}

// SetQueryParam sets a URL query param to a given value.
func (r *Request) SetQueryParam(param, value string) {
	q := r.URL.Query()
	q.Set(param, value)
	r.Request.URL.RawQuery = q.Encode()
}

// SetURLPathArgument sets a URL path argument to a given value.
func (r *Request) SetURLPathArgument(arg, value string) {
	path := strings.Split(r.URL.EscapedPath(), "/")
	for index, item := range path {
		if arg == item {
			path[index] = value
		}
	}

	r.Request.URL.Path = strings.Join(path, "/")
}

// Response is a *http.Response that allows cloning its body.
type Response struct {
	*http.Response
}

// CloneBody makes a copy of a response, including its body, while leaving the original body intact.
func (r *Response) CloneBody() (*Response, error) {
	newResponse := new(http.Response)

	if r.Response.Header != nil {
		newResponse.Header = r.Response.Header.Clone()
	}

	if r.Response.Trailer != nil {
		newResponse.Trailer = r.Response.Trailer.Clone()
	}

	newResponse.ContentLength = r.Response.ContentLength
	newResponse.Uncompressed = r.Response.Uncompressed
	newResponse.Request = r.Response.Request
	newResponse.TLS = r.Response.TLS
	newResponse.Status = r.Response.Status
	newResponse.StatusCode = r.Response.StatusCode
	newResponse.Proto = r.Response.Proto
	newResponse.ProtoMajor = r.Response.ProtoMajor
	newResponse.ProtoMinor = r.Response.ProtoMinor
	newResponse.Close = r.Response.Close
	copy(newResponse.TransferEncoding, r.Response.TransferEncoding)

	if r.Response.Body == nil {
		return &Response{Response: newResponse}, nil
	}

	body, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return &Response{Response: newResponse}, err
	}

	// Put back the original body
	r.Response.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Clone the request body
	newResponse.Body = ioutil.NopCloser(bytes.NewReader(body))
	return &Response{Response: newResponse}, nil
}
