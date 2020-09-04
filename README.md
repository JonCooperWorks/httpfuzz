httpfuzz
--------

[![PkgGoDev](https://pkg.go.dev/badge/github.com/joncooperworks/httpfuzz)](https://pkg.go.dev/github.com/joncooperworks/httpfuzz)

`httpfuzz` is a fast HTTP fuzzer written in [Go](https://golang.org) inspired by [Burp Intruder](https://portswigger.net/burp/documentation/desktop/tools/intruder).
It takes a seed request and uses a wordlist to generate requests.
For a wordlist with `m` words and a seed request with `n` injection points, `httpfuzz` will generate `m * n` requests.
It can be used as a library, but is meant to be used with the included `httpfuzz` CLI.

## Using httpfuzz CLI
`
   httpfuzz - fuzz endpoints based on a HTTP request file

USAGE:
   httpfuzz [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --count-only              don't send the requests, just count how many would be sent (default: false)
   --seed-request value      the request to be fuzzed
   --delay-ms value          the delay between each HTTP request in milliseconds (default: 0)]
   --wordlist value          newline separated wordlist for the fuzzer
   --target-header value     HTTP headers to fuzz
   --https                   (default: false)
   --skip-cert-verify        skip verifying SSL certificate when making requests (default: false)
   --proxy-url value         HTTP proxy to send requests through
   --proxy-ca-pem value      PEM encoded CA Certificate for TLS requests through a proxy
   --target-param value      URL Query string param to fuzz
   --target-path-arg value   URL path argument to fuzz
   --dirbuster               brute force directory names from wordlist (default: false)
   --target-delimiter value  delimiter to mark targets in request bodies (default: "`")
   --help, -h                show help (default: false)
`

Seed requests are a text HTTP request.
You can tag injection points in request bodies by surrounding them with the delimiter character specified at program startup with the `--target-delimiter` flag.
By default, it's `` ` ``.
You can fuzz other parts of the request by with CLI flags.

### Examples

`
POST /api/devices HTTP/1.1
Content-Type: application/json
User-Agent: PostmanRuntime/7.26.3
Accept: */*
Cache-Control: no-cache
Postman-Token: c5bcc2bc-90b4-4d06-b851-1cc670cd9afa
Host: localhost:8000
Accept-Encoding: gzip, deflate
Connection: close
Content-Length: 35

{
	"name": "`S9`",
	"os": "Android"
}
`

The backticks (`` ` ``) indicate a spot in the request body to inject payloads from the wordlist.

`
httpfuzz \
   --wordlist testdata/useragents.txt \
   --seed-request testdata/validPOST.request \
   --target-header User-Agent \
   --target-header Host \
   --delay-ms 50 \
   --target-header Pragma \
   --skip-cert-verify \
   --proxy-url http://localhost:8080 \
   --target-param fuzz \
   --https \
   --dirbuster
`

In the above example, `httpfuzz` will insert values from the wordlist into the `name` field, the `Pragma`, `User-Agent` and `Host` headers, the end of the URL (like [dirbuster](https://tools.kali.org/web-applications/dirbuster#:~:text=DirBuster%20is%20a%20multi%20threaded,pages%20and%20applications%20hidden%20within.)) and the URL parameter `fuzz`.

### Building httpfuzz
To build `httpfuzz`, simply run `go build -o httpfuzz cmd/httpfuzz.go`.
You can run the tests with `go test -v`.