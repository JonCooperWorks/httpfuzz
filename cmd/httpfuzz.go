package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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
	transport := &http.Transport{}
	if c.Bool("skip-cert-verify") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	proxyURL := c.String("proxy-url")
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return err
		}

		transport.Proxy = http.ProxyURL(proxy)
	}

	httpClient.Transport = transport
	logger := log.New(os.Stdout, "httpfuzz: ", log.Ldate|log.Ltime|log.Lshortfile)

	var urlScheme string
	if c.Bool("https") {
		urlScheme = "https"
	} else {
		urlScheme = "http"
	}

	config := &httpfuzz.Config{
		TargetHeaders: c.StringSlice("target-header"),
		Wordlist:      wordlist,
		Client:        &httpfuzz.Client{Client: httpClient},
		Seed:          &httpfuzz.Request{Request: request},
		Logger:        logger,
		RequestDelay:  time.Duration(c.Int("delay-ms")) * time.Millisecond,
		URLScheme:     urlScheme,
	}

	fuzzer := &httpfuzz.Fuzzer{Config: config}
	requestCount, err := fuzzer.RequestCount()
	if err != nil {
		return err
	}

	logger.Printf("Sending %d requests", requestCount)

	if !c.Bool("count-only") {
		fuzzer.WaitFor(requestCount)
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
			&cli.IntFlag{
				Name:     "delay-ms",
				Required: false,
				Usage:    "the delay between each HTTP request in milliseconds",
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
			&cli.BoolFlag{
				Name:     "https",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "skip-cert-verify",
				Required: false,
				Value:    false,
				Usage:    "skip verifying SSL certificate when making requests",
			},
			&cli.StringFlag{
				Name:     "proxy-url",
				Required: false,
				Usage:    "HTTP proxy to send requests through",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
