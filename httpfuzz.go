// Package httpfuzz is a fast fuzzer that allows you to easily fuzz HTTP endpoints.
// It works in a similar way to Burp Intruder, but it doesn't read the entire wordlist into memory.
// Instead, it calculates how many requests it's going to send ahead of time and streams through the wordlist line-by-line, using go's sync.WaitGroup to wait until the last request finishes.
package httpfuzz
