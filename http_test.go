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

func TestSetDirectoryRoot(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("body"))
	request := &Request{req}
	request.SetDirectoryRoot("added")

	expectedURL := "/test/path/added?param=test"
	actualURL := request.URL.String()
	if actualURL != expectedURL {
		t.Fatalf("Expected %s, got %s", expectedURL, actualURL)
	}
}

func TestBodyTargetCount(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("*body**second*"))
	request := &Request{req}

	count, err := request.BodyTargetCount('*')
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Fatalf("Expected 2, got %d", count)
	}
}

func TestBodyTargetCountUnbalancedDelimiters(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("*body"))
	request := &Request{req}

	count, err := request.BodyTargetCount('*')
	if err == nil {
		t.Fatalf("Expected error, got %d", count)
	}

	if count != 0 {
		t.Fatalf("Expected 0 count, got %d", count)
	}
}

func TestRemoveDelimiters(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("{\"type\": \"*body*\", \"second\": \"*value*\"}"))
	request := &Request{req}
	previousContentLength := request.ContentLength
	targetCount, _ := request.BodyTargetCount('*')
	err := request.RemoveDelimiters('*')
	if err != nil {
		t.Fatal(err)
	}

	expectedContentLength := previousContentLength - int64(targetCount*2)
	if request.ContentLength != expectedContentLength {
		t.Fatalf("Content length does not match, expected %d, got %d", expectedContentLength, request.ContentLength)
	}

	expectedBody := []byte("{\"type\": \"body\", \"second\": \"value\"}")
	actualBody, _ := ioutil.ReadAll(request.Body)

	// Ensure request body is not consumed when replacing delimiters.
	if len(actualBody) == 0 {
		t.Fatal("Request body was consumed")
	}

	if !bytes.Equal(actualBody, expectedBody) {
		t.Fatalf("bodies do not match. expected %s, got %s", string(expectedBody), string(actualBody))
	}
}

func TestRemoveDelimitersEmptyRequestBody(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test/path?param=test", nil)
	request := &Request{req}
	err := request.RemoveDelimiters('*')
	if err != nil {
		t.Fatal(err)
	}

	// Ensure request body remains empty.
	if request.Body != nil {
		t.Fatal("Request body not expected")
	}
}

func TestInjectPayload(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("{\"type\": \"*body*\", \"second\": \"*value*\"}"))
	request := &Request{req}
	err := request.SetBodyPayloadAt(0, '*', "test")
	if err != nil {
		t.Fatal(err)
	}

	expectedBody := []byte("{\"type\": \"test\", \"second\": \"*value*\"}")
	actualBody, _ := ioutil.ReadAll(request.Body)

	expectedContentLength := int64(len(expectedBody))
	if request.ContentLength != expectedContentLength {
		t.Fatalf("Content length does not match, expected %d, got %d", expectedContentLength, request.ContentLength)
	}

	// Ensure request body is not consumed when injecting payload.
	if len(actualBody) == 0 {
		t.Fatal("Request body was consumed")
	}

	if !bytes.Equal(actualBody, expectedBody) {
		t.Fatalf("bodies do not match. expected %s, got %s", string(expectedBody), string(actualBody))
	}
}

func TestInjectPayloadUnbalancedDelimiters(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test/path?param=test", strings.NewReader("{\"type\": \"*body\", \"second\": \"*value*\"}"))
	request := &Request{req}
	err := request.SetBodyPayloadAt(0, '*', "test")
	if err == nil {
		t.Fatal("Expected error with imbalanced delimiters.")
	}
}
