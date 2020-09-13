package httpfuzz

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
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

// IsMultipartForm returns true if this is a multipart request.
func (r *Request) IsMultipartForm() bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return false
	}

	return strings.HasPrefix(mediaType, "multipart/")
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
	defer req.Body.Close()

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

// SetDirectoryRoot inserts a string after the final "/" in a URL to
func (r *Request) SetDirectoryRoot(value string) {
	path := strings.Split(r.URL.EscapedPath(), "/")
	path = append(path, value)
	r.Request.URL.Path = strings.Join(path, "/")
}

// BodyTargetCount calculates the number of targets in a request body.
func (r *Request) BodyTargetCount(delimiter byte) (int, error) {
	if r.Body == nil {
		return 0, nil
	}

	clone, err := r.CloneBody(context.Background())
	if err != nil {
		return 0, err
	}

	count := 0
	buf := make([]byte, bufio.MaxScanTokenSize)

	for {
		bufferSize, err := clone.Body.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], delimiter)
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

	if count%2 != 0 {
		return 0, fmt.Errorf("unbalanced delimiters")
	}

	return count / 2, nil
}

// RemoveDelimiters removes all target delimiters from a request so it can be sent to the server and interpreted properly.
func (r *Request) RemoveDelimiters(delimiter byte) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}

	// Prevent accidentally mangling multipart requests
	if r.IsMultipartForm() {
		return nil
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body = bytes.Replace(body, []byte{delimiter}, []byte{}, -1)

	// Adjust content length
	r.Request.ContentLength = int64(len(body))

	// Put back request body without the delimiters.
	r.Request.Body = ioutil.NopCloser(bytes.NewReader(body))
	return nil
}

// SetBodyPayloadAt injects a payload at a given position.
func (r *Request) SetBodyPayloadAt(position int, delimiter byte, payload string) error {
	if r.Body == nil {
		return nil
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Calculate the offsets in the body that correspond to the position
	index := &DelimiterArray{Contents: body}
	delimiterPositions := index.Lookup(byte(delimiter))
	if len(delimiterPositions)%2 != 0 {
		return fmt.Errorf("unbalanced delimiters")
	}
	start, end, err := delimiterIndex(position, delimiterPositions)
	if err != nil {
		return err
	}

	// Replace bytes between the start and end offset with payload bytes and the delimiters removed.
	prefix := body[0:start]
	suffix := body[end+1 : len(body)]
	newBody := []byte{}
	newBody = append(newBody, prefix...)
	newBody = append(newBody, []byte(payload)...)
	newBody = append(newBody, suffix...)

	// Adjust content length
	r.Request.ContentLength = int64(len(newBody))

	// Put back request body with the injected target.
	r.Request.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	return nil
}

// ReplaceMultipartFileData replaces a file in the request body with a generated payload.
func (r *Request) ReplaceMultipartFileData(fieldName string, file *File) error {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	if !r.IsMultipartForm() {
		return fmt.Errorf("request is not a multipart request, got %s", mediaType)
	}

	boundary := params["boundary"]
	mr := multipart.NewReader(r.Body, boundary)
	newBody := &bytes.Buffer{}

	// Write from the old writer into the new one using the same boundary as the original request.
	written := false
	mw := multipart.NewWriter(newBody)
	mw.SetBoundary(boundary)
	var chunk []byte
	for {
		defer mw.Close()
		part, err := mr.NextPart()
		if err == io.EOF {
			if !written {
				partWriter, err := mw.CreatePart(part.Header)
				_, err = partWriter.Write(file.Payload)
				if err != nil {
					return err
				}
			}
			break
		}
		if err != nil {
			return err
		}

		chunk, err = ioutil.ReadAll(part)
		if err != nil {
			return err
		}

		_, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil {
			return err
		}

		// Copy part headers from the old request
		if params["name"] == fieldName {
			partWriter, err := mw.CreatePart(part.Header)
			_, err = partWriter.Write(file.Payload)
			if err != nil {
				return err
			}
			written = true
		} else {
			partWriter, err := mw.CreatePart(part.Header)
			if err != nil {
				return err
			}

			_, err = partWriter.Write(chunk)
			if err != nil {
				return err
			}
		}
	}

	r.ContentLength = int64(newBody.Len()) + 130
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Body = ioutil.NopCloser(newBody)
	return nil
}

// ReplaceMultipartField replaces a regular form field in a multipart request with a payload.
// We do this because delimiters don't work with binary files.
func (r *Request) ReplaceMultipartField(fieldName, payload string) error {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	if !r.IsMultipartForm() {
		return fmt.Errorf("request is not a multipart request, got %s", mediaType)
	}

	mr := multipart.NewReader(r.Body, params["boundary"])
	newBody := &bytes.Buffer{}

	written := false

	// Write from the old writer into the new one using the same boundary as the original request.
	mw := multipart.NewWriter(newBody)
	mw.SetBoundary(params["boundary"])
	for {
		defer mw.Close()
		part, err := mr.NextPart()
		if err == io.EOF {
			if !written {
				writer, err := mw.CreateFormField(fieldName)
				if err != nil {
					return err
				}
				_, err = writer.Write([]byte(payload))
				if err != nil {
					return err
				}
			}
			break
		}
		if err != nil {
			return err
		}

		chunk, err := ioutil.ReadAll(part)
		if err != nil {
			return err
		}

		_, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil {
			return err
		}

		// Copy part headers from the old request
		if params["name"] == fieldName {
			r.Request.ContentLength = r.ContentLength - (int64(len(chunk)) - int64(len(payload)))
			writer, err := mw.CreateFormField(fieldName)
			if err != nil {
				return err
			}
			_, err = writer.Write([]byte(payload))
			if err != nil {
				return err
			}
			written = true
		} else {
			partWriter, err := mw.CreatePart(part.Header)
			if err != nil {
				return err
			}

			_, err = partWriter.Write(chunk)
			if err != nil {
				return err
			}
		}
	}

	r.ContentLength = int64(newBody.Len()) + 130
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.Body = ioutil.NopCloser(newBody)
	return nil
}

func delimiterIndex(position int, delimiterPositions []int) (int, int, error) {
	for i := 0; i < len(delimiterPositions); i++ {
		if i/2-position <= 1 {
			return delimiterPositions[i], delimiterPositions[i+1], nil
		}
	}

	return 0, 0, fmt.Errorf("position out of range")
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
	defer r.Response.Body.Close()

	// Put back the original body
	r.Response.Body = ioutil.NopCloser(bytes.NewReader(body))

	// Clone the request body
	newResponse.Body = ioutil.NopCloser(bytes.NewReader(body))
	return &Response{Response: newResponse}, nil
}
