package httpfuzz

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
)

// Wordlist is a stream of the words in httpfuzz's wordlist.
type Wordlist struct {
	File *os.File
	mux  sync.Mutex
}

// Stream returns a <- chan string that receives lines as they come from the wordlist file.
// It does not rewind the file after using it.
func (w *Wordlist) Stream() <-chan string {
	payloads := make(chan string)

	// Ensure only one stream can run at a time per wordlist.
	w.mux.Lock()
	go func(payloads chan<- string) {
		// If there is no wordlist, just close the chan.
		if w.File == nil {
			close(payloads)
			return
		}

		defer w.mux.Unlock()
		scanner := bufio.NewScanner(w.File)
		for scanner.Scan() {
			payloads <- scanner.Text()
		}
		close(payloads)
	}(payloads)
	return payloads
}

// Count returns the number of words in a wordlist.
func (w *Wordlist) Count() (int, error) {
	// If there's no wordlist, there are no files in it.
	if w.File == nil {
		return 0, nil
	}

	// We don't want to start a count in the middle of a stream.
	w.mux.Lock()
	defer w.mux.Unlock()
	count := 1
	const lineBreak = '\n'

	buf := make([]byte, bufio.MaxScanTokenSize)

	for {
		bufferSize, err := w.File.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}

		var buffPosition int
		for {
			i := bytes.IndexByte(buf[buffPosition:], lineBreak)
			if i == -1 || bufferSize == buffPosition {
				break
			}
			buffPosition += i + 1
			count++
		}
		if err == io.EOF {
			break
		}
	}

	// Move back to the head of the file
	_, err := w.File.Seek(0, io.SeekStart)
	if err != nil {
		return count, err
	}

	return count, nil
}
