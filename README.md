httpfuzz
--------

[![PkgGoDev](https://pkg.go.dev/badge/github.com/joncooperworks/httpfuzz)](https://pkg.go.dev/github.com/joncooperworks/httpfuzz)

`httpfuzz` is a fast HTTP fuzzer written in [Go](https://golang.org) inspired by [Burp Intruder](https://portswigger.net/burp/documentation/desktop/tools/intruder).
It takes a seed request and uses a wordlist to generate requests.
For a wordlist with `m` words and a seed request with `n` injection points, `httpfuzz` will generate `m * n` requests.
It can be used as a library, but is meant to be used with the included `httpfuzz` CLI.
It allows fuzzing of HTTP requests with text bodies and multipart file uploads.

### File Fuzzing
`httpfuzz` can generate files to help you quickly test a file upload endpoints for file header whitelisting using the `--automatic-file-payloads` flag.
It generates random bytes and puts a valid file headers on them before injecting them into the request body and sending them to the web service.
This lets you easily see if a dev team is only using filenames to validate image uploads ;).
You can see a list of the supported file types in [fileheaders.go](https://github.com/JonCooperWorks/httpfuzz/blob/master/fileheaders.go).
Feel free to add any file types you see missing.

If you want to use `httpfuzz` with existing payloads, simply place them in a directory and pass it to the `payload-dir` flag.

## Using httpfuzz CLI
```
NAME:
   httpfuzz - fuzz endpoints based on a HTTP request file

USAGE:
   httpfuzz [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --count-only                 don't send the requests, just count how many would be sent (default: false)
   --seed-request value         the request to be fuzzed
   --delay-ms value             the delay between each HTTP request in milliseconds (default: 0)
   --wordlist value             newline separated wordlist for the fuzzer
   --target-header value        HTTP headers to fuzz
   --https                      (default: false)
   --skip-cert-verify           skip verifying SSL certificate when making requests (default: false)
   --proxy-url value            HTTP proxy to send requests through
   --proxy-ca-pem value         PEM encoded CA Certificate for TLS requests through a proxy
   --target-param value         URL Query string param to fuzz
   --target-path-arg value      URL path argument to fuzz
   --dirbuster                  brute force directory names from wordlist (default: false)
   --target-delimiter value     delimiter to mark targets in request bodies (default: "`")
   --multipart-file-name value  name of the file field to fuzz in multipart request
   --multipart-form-name value  name of the form field to fuzz in multipart request
   --fuzz-file-size value       file size to fuzz in multipart request (default: 1024)
   --payload-dir value          directory with payload files to attempt to upload using the fuzzer
   --automatic-file-payloads    enable this flag to automatically generate files for fuzzing (default: false)
   --help, -h                   show help (default: false)
```

Seed requests are a text HTTP request.
You can tag injection points in request bodies by surrounding them with the delimiter character specified at program startup with the `--target-delimiter` flag.
By default, it's `` ` ``.
You can fuzz other parts of the request with CLI flags.

### Examples

#### Fuzzing POST requests

```
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
```

The backticks (`` ` ``) indicate a spot in the request body to inject payloads from the wordlist.

```
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
   --dirbuster
```

In the above example, `httpfuzz` will insert values from the wordlist into the `name` field, the `Pragma`, `User-Agent` and `Host` headers, the end of the URL (like [dirbuster](https://tools.kali.org/web-applications/dirbuster#:~:text=DirBuster%20is%20a%20multi%20threaded,pages%20and%20applications%20hidden%20within.)) and the URL parameter `fuzz`.

#### Fuzzing multipart file uploads
```
POST /uploadFile HTTP/1.1
Host: localhost:8000
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0
Content-Length: 1309
Accept: application/json
Accept-Encoding: gzip, deflate
Accept-Language: en-US,en;q=0.5
Cache-Control: no-cache
Connection: close
Content-Type: multipart/form-data; boundary=---------------------------416891666813988703772682177556
X-Requested-With: XMLHttpRequest

-----------------------------416891666813988703772682177556
Content-Disposition: form-data; name="file"; filename="image.png"
Content-Type: image/png

*image data here. real request is in validuploadPOST.request*
-----------------------------416891666813988703772682177556--

-----------------------------416891666813988703772682177556--
```

That request uploads a PNG file.
You can fuzz it with the following command:
```
httpfuzz \
 --wordlist testdata/useragents.txt \
 --seed-request testdata/validuploadPOST.request \
 --target-header User-Agent \
 --target-header Host \
 --delay-ms 50 \
 --target-header Pragma \
 --proxy-url http://localhost:8080 \
 --target-param fuzz \
 --dirbuster \
 --fuzz-file-size 4096 \
 --multipart-form-name field \
 --multipart-file-name file \
 --automatic-file-payloads \
 --payload-dir ./testpayloads
```

This command will fuzz a multipart form field called `field` and the file field `file` with randomly generated 4KB (4096 bytes) files and any payloads in the `./testpayloads` directory.
You can still fuzz the other injection points, but delimiter injection will not work, since binary files can contain any character they want.

### Building httpfuzz
To build `httpfuzz`, simply run `go build -o httpfuzz cmd/httpfuzz.go`.
You can run the tests with `go test -v`.