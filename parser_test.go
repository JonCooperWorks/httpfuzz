package httpfuzz

import (
	"io/ioutil"
	"testing"
)

func TestHTTPRequestInvalidFileReturnsError(t *testing.T) {
	req, err := RequestFromFile("notfound.request")
	if err == nil {
		t.Fatalf("expected error")
	}

	if req != nil {
		t.Fatalf("request returned when expected nil: %+v", req)
	}
}

func TestHTTPRequestParsedCorrectlyFromFile(t *testing.T) {
	req, err := RequestFromFile("./testdata/validGET.request")
	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	if req.Method != "GET" {
		t.Fatalf("expected GET, got %v", req.Method)
	}

	if req.Host != "localhost:8000" {
		t.Fatalf("expected URL 'localhost:8000', got %v", req.Host)
	}

	if req.UserAgent() != "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0) Gecko/20100101 Firefox/78.0" {
		t.Fatalf("got unexpected User-Agent %v", req.UserAgent())
	}

	if req.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("got unexpected cache-control header")
	}
}

func TestPOSTRequestBodyParsedCorrectlyFromFile(t *testing.T) {
	req, err := RequestFromFile("./testdata/validPOST.request")
	if err != nil {
		t.Fatalf("expected err to be nil, got %v", err)
	}

	if req.Method != "POST" {
		t.Fatalf("expected POST, got %v", req.Method)
	}

	if req.Host != "localhost:8000" {
		t.Fatalf("expected URL 'localhost:8000', got %v", req.Host)
	}

	if req.UserAgent() != "PostmanRuntime/7.26.3" {
		t.Fatalf("got unexpected User-Agent %v", req.UserAgent())
	}

	if req.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("got unexpected cache-control header")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Check for truncation bug
	const expectedLength = 41
	if len(body) != expectedLength {
		t.Log(string(body))
		t.Fatalf("Body is incorrect length, expected %d, got %d", expectedLength, len(body))
	}
}
