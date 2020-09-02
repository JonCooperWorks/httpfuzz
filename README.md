httpfuzz
--------
```httpfuzz``` is a fast HTTP fuzzer written in [Go](https://golang.org).
It's inspired by [Burp Intruder](https://portswigger.net/burp/documentation/desktop/tools/intruder).
It takes a seed request and uses a wordlist to generate requests.
For ```m``` requests and ```n``` injection points, it will generate ```m * n``` requests.

## Using httpfuzz
```
   httpfuzz - fuzz endpoints based on a HTTP request file

USAGE:
   httpfuzz [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --count-only              don't send the requests, just count how many would be sent (default: false)
   --seed-request value      the request to be fuzzed
   --delay-ms value          the delay between each HTTP request in milliseconds (default: 0)
   --plugin value            httpfuzz plugin binary
   --plugin-arg value        httpfuzz plugin argument. in the form plugin:value
   --wordlist value          newline separated wordlist for the fuzzer
   --target-header value     HTTP headers to fuzz
   --https                   (default: false)
   --skip-cert-verify        skip verifying SSL certificate when making requests (default: false)
   --proxy-url value         HTTP proxy to send requests through
   --proxy-ca-pem value      PEM encoded CA Certificate for TLS requests through a proxy
   --target-param value      URL Query string param to fuzz
   --target-path-arg value   URL path argument to fuzz
   --dirbuster               brute force directory names from wordlist (default: false)
   --target-delimiter value  delimiter to mark targets in request bodies (default: "*")
   --help, -h                show help (default: false)
```