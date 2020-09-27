package httpfuzz

import (
	"log"
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
	TargetFilenames           []string
	LogSuccess                bool
	EnableGeneratedPayloads   bool
	FuzzFileSize              int64
	FuzzDirectory             bool
	Wordlist                  *Wordlist
	Seed                      *Request
	Client                    *Client
	RequestDelay              time.Duration
	Plugins                   *PluginBroker
	Logger                    *log.Logger
	URLScheme                 string
	TargetDelimiter           byte
	waitGroup                 sync.WaitGroup
}
