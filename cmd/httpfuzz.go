package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joncooperworks/httpfuzz"
	"github.com/urfave/cli/v2"
)

func actionHTTPFuzz(c *cli.Context) error {
	wordlist, err := os.Open(c.String("wordlist"))
	if err != nil {
		return err
	}
	defer wordlist.Close()

	request, err := httpfuzz.RequestFromFile(c.String("seed-request"))
	if err != nil {
		return err
	}

	httpClient := &http.Client{}
	logger := log.New(os.Stdout, "httpfuzz", log.Llongfile)

	config := &httpfuzz.Config{
		TargetHeaders:         c.StringSlice("target-header"),
		Wordlist:              wordlist,
		Client:                &httpfuzz.Client{Client: httpClient},
		Seed:                  &httpfuzz.Request{Request: request},
		MaxConcurrentRequests: c.Int64("max-concurrent-requests"),
		Logger:                logger,
		URLScheme:             c.String("url-scheme"),
	}

	fuzzer := &httpfuzz.Fuzzer{Config: config}
	requestCount, err := fuzzer.RequestCount()
	if err != nil {
		return err
	}

	logger.Printf("Sending %d requests", requestCount)

	_, err = wordlist.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	if !c.Bool("count-only") {
		config.WaitGroup.Add(requestCount)
		requests := fuzzer.GenerateRequests()
		fuzzer.ProcessRequests(requests)
		logger.Printf("Finished.")
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:   "httpfuzz",
		Usage:  "fuzz endpoints based on a HTTP request file",
		Action: actionHTTPFuzz,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "count-only",
				Required: false,
				Usage:    "don't send the requests, just count how many would be sent",
			},
			&cli.StringFlag{
				Name:     "seed-request",
				Required: true,
				Usage:    "the request to be fuzzed",
			},
			&cli.Int64Flag{
				Name:     "max-concurrent-requests",
				Required: false,
				Usage:    "the number of requests to run at once",
			},
			&cli.StringFlag{
				Name:     "plugins-dir",
				Required: false,
				Usage:    "directory for plugin binaries",
			},
			&cli.StringFlag{
				Name:     "wordlist",
				Required: true,
				Usage:    "newline separated wordlist for the fuzzer",
			},
			&cli.StringSliceFlag{
				Name:     "target-header",
				Required: false,
				Usage:    "HTTP headers to fuzz",
			},
			&cli.StringFlag{
				Name:     "url-scheme",
				Required: false,
				Value:    "http",
				Usage:    "URL scheme for requests. http or https",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
