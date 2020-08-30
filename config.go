package httpfuzz

import (
	"log"
	"os"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders         []string
	Wordlist              *os.File
	Seed                  *Request
	Client                *Client
	MaxConcurrentRequests int64
	Plugins               []Plugin
	Logger                *log.Logger
}
