package httpfuzz

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
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

// Request is a *http.Request that allows cloning its body.
type Request struct {
	*http.Request
}

// CloneBody makes a copy of a request, including its body, while leaving the original body intact.
func (r *Request) CloneBody(ctx context.Context) *Request {
	req := &Request{Request: r.Request.Clone(ctx)}
	if req.Body == nil {
		return req
	}

	body, _ := ioutil.ReadAll(req.Body)

	// Put back the original body
	r.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Clone the request body
	req.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
	return req
}

// Response is a *http.Response that allows cloning its body.
type Response struct {
	*http.Response
}
