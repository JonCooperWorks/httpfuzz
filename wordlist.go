package httpfuzz

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

// Wordlist is a stream of the words in httpfuzz's wordlist.
type Wordlist struct {
	File *os.File
}

// Stream returns a <- chan string that receives lines as they come from the wordlist file.
func (w *Wordlist) Stream() <-chan string {
	payloads := make(chan string)
	go func(payloads chan<- string) {
		scanner := bufio.NewScanner(w.File)
		for scanner.Scan() {
			payload := scanner.Text()
			payloads <- payload
		}
		close(payloads)
	}(payloads)
	return payloads
}

// Count returns the number of words in a wordlist.
func (w *Wordlist) Count() (int, error) {
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
