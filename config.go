package httpfuzz

import (
	"log"
	"os"
	"sync"
	"time"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders  []string
	TargetParams   []string
	TargetPathArgs []string
	FuzzDirectory  bool
	Wordlist       *os.File
	Seed           *Request
	Client         *Client
	RequestDelay   time.Duration
	Plugins        []Plugin
	Logger         *log.Logger
	URLScheme      string
	waitGroup      sync.WaitGroup
}
