package httpfuzz

import (
	"io"
	"log"
	"sync"
	"time"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders []string
	Wordlist      io.Reader
	Seed          *Request
	Client        *Client
	RequestDelay  time.Duration
	Plugins       []Plugin
	Logger        *log.Logger
	URLScheme     string
	waitGroup     sync.WaitGroup
}
