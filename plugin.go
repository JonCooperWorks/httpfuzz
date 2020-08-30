package httpfuzz

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	OnError(req *Request, err error)
	OnSuccess(req *Request, resp *Response)
}
