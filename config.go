package httpfuzz

import (
	"log"
	"os"
	"sync"
	"time"
)

// Config holds all fuzzer configuration.
type Config struct {
	TargetHeaders             []string
	TargetParams              []string
	TargetPathArgs            []string
	TargetFileKeys            []string
	TargetMultipartFieldNames []string
	FilesystemPayloads        []string
	EnableGeneratedPayloads   bool
	FuzzFileSize              int64
	FuzzDirectory             bool
	Wordlist                  *os.File
	Seed                      *Request
	Client                    *Client
	RequestDelay              time.Duration
	Plugins                   []Plugin
	Logger                    *log.Logger
	URLScheme                 string
	TargetDelimiter           byte
	waitGroup                 sync.WaitGroup
}
