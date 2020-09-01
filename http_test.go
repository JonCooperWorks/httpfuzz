package httpfuzz

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestRequestClonePreservesOriginalBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "", strings.NewReader("body"))
	request := &Request{req}
	clonedRequest, err := request.CloneBody(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Read the cloned body first to make sure the original body doesn't get consumed.
	clonedBody, _ := ioutil.ReadAll(clonedRequest.Body)
	body, _ := ioutil.ReadAll(request.Body)

	if len(body) == 0 {
		t.Fatalf("Original body was drained with the clone")
	}

	if !bytes.Equal(body, clonedBody) {
		t.Fatalf("Cloned body does not match original, expected %s, got %s", string(body), string(clonedBody))
	}

	if req.RequestURI != "" {
		t.Fatalf("RequestURI was not removed before clone")
	}

	if request.Host != clonedRequest.URL.Host {
		t.Fatalf("Host was not copied into URL")
	}
}

func TestResponseClonePreservesOriginalBody(t *testing.T) {
	body := "body"
	resp := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        http.Header{},
		Body:          ioutil.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	response := &Response{resp}
	clonedResponse, err := response.CloneBody()
	if err != nil {
		t.Fatal(err)
	}

	// Read the cloned body first to make sure the original body doesn't get consumed.
	clonedBody, _ := ioutil.ReadAll(clonedResponse.Body)
	originalBody, _ := ioutil.ReadAll(response.Body)

	if len(originalBody) == 0 {
		t.Fatal("Original body was consumed when clone body was read")
	}

	if !bytes.Equal(clonedBody, originalBody) {
		t.Fatalf("Cloned body does not match original, expected %s, got %s", string(originalBody), string(clonedBody))
	}

	if response.Status != clonedResponse.Status {
		t.Fatalf("Cloned response status does not match original response status")
	}
}

func TestHasPathArgument(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path", strings.NewReader("body"))
	request := &Request{req}
	if !request.HasPathArgument("path") {
		t.Fatal("Expected HasPathArgument to be true")
	}

	if request.HasPathArgument("notfound") {
		t.Fatalf("Expected HasPathArgument to be false")
	}
}

func TestSetQueryParam(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path", strings.NewReader("body"))
	request := &Request{req}
	request.SetQueryParam("param", "test")

	expectedURL := "/test/path?param=test"
	actualURL := request.URL.String()
	if actualURL != expectedURL {
		t.Fatalf("Expected %s, got %s", expectedURL, actualURL)
	}
}

func TestSetURLPathArgument(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("body"))
	request := &Request{req}
	request.SetURLPathArgument("path", "test")

	expectedURL := "/test/test?param=test"
	actualURL := request.URL.String()
	if actualURL != expectedURL {
		t.Fatalf("Expected %s, got %s", expectedURL, actualURL)
	}
}
