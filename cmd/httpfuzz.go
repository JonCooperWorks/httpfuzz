package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
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

	seedRequest, err := httpfuzz.RequestFromFile(c.String("seed-request"))
	if err != nil {
		return err
	}

	targetPathArgs := c.StringSlice("target-path-arg")
	for _, arg := range targetPathArgs {
		if !seedRequest.HasPathArgument(arg) {
			return fmt.Errorf("seed request does not have URL path arg '%s'", arg)
		}
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if proxyCACertFilename := c.String("proxy-ca-pem"); proxyCACertFilename != "" {
		proxyCACertFile, err := os.Open(proxyCACertFilename)
		if err != nil {
			return err
		}
		defer proxyCACertFile.Close()

		certs, err := ioutil.ReadAll(proxyCACertFile)
		if err != nil {
			return err
		}

		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			return fmt.Errorf("failed to trust custom CA certs from %s", proxyCACertFilename)
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Bool("skip-cert-verify"),
			RootCAs:            rootCAs,
		},
	}

	if proxyURL := c.String("proxy-url"); proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return err
		}

		transport.Proxy = http.ProxyURL(proxy)
	}

	httpClient := &http.Client{
		Transport: transport,
	}
	logger := log.New(os.Stdout, "httpfuzz: ", log.Ldate|log.Ltime|log.Lshortfile)

	var urlScheme string
	if c.Bool("https") {
		urlScheme = "https"
	} else {
		urlScheme = "http"
	}

	plugins, err := httpfuzz.LoadPlugins(logger, c.StringSlice("plugin"), c.StringSlice("plugin-arg"))
	if err != nil {
		return err
	}

	delimiter := []byte(c.String("target-delimiter"))[0]

	// Validate that the request body is properly delimitered
	_, err = seedRequest.BodyTargetCount(delimiter)
	if err != nil {
		return err
	}

	client := &httpfuzz.Client{
		Client: httpClient,
	}

	config := &httpfuzz.Config{
		TargetHeaders:   c.StringSlice("target-header"),
		TargetParams:    c.StringSlice("target-param"),
		FuzzDirectory:   c.Bool("dirbuster"),
		TargetPathArgs:  targetPathArgs,
		Wordlist:        wordlist,
		Client:          client,
		Seed:            seedRequest,
		TargetDelimiter: delimiter,
		Logger:          logger,
		RequestDelay:    time.Duration(c.Int("delay-ms")) * time.Millisecond,
		URLScheme:       urlScheme,
		Plugins:         plugins,
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
			&cli.StringFlag{
				Name:     "proxy-ca-pem",
				Required: false,
				Usage:    "PEM encoded CA Certificate for TLS requests through a proxy",
			},
			&cli.StringSliceFlag{
				Name:  "target-param",
				Usage: "URL Query string param to fuzz",
			},
			&cli.StringSliceFlag{
				Name:  "target-path-arg",
				Usage: "URL path argument to fuzz",
			},
			&cli.BoolFlag{
				Name:     "dirbuster",
				Required: false,
				Usage:    "brute force directory names from wordlist",
			},
			&cli.StringFlag{
				Name:  "target-delimiter",
				Usage: "delimiter to mark targets in request bodies",
				Value: "`",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
