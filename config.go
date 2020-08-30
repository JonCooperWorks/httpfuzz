package httpfuzz

import (
	"os"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders         []string
	Wordlist              *os.File
	Seed                  *Request
	Client                *Client
	MaxConcurrentRequests int
	Plugins               []Plugin
}
