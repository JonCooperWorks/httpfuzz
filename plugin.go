package httpfuzz

import "net/http"

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	OnError(req *http.Request, err error)
	OnSuccess(req *http.Request, resp *http.Response)
}
