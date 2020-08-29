package httpfuzz

import (
	"net/http"
	"os"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaderWordlists map[string]*os.File
	URLPathWordlist       *os.File
	Seed                  *http.Request
	Client                *http.Client
	MaxConcurrentRequests int
	Plugins               []Plugin
}
