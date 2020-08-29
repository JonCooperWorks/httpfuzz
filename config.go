package httpfuzz

import (
	"net/http"
	"os"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders         []string
	FuzzURL               bool
	Wordlist              *os.File
	Seed                  *http.Request
	Client                *http.Client
	MaxConcurrentRequests int
	Plugins               []Plugin
}
