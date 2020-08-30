package httpfuzz

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	OnSuccess(req *Request, resp *Response) error
	Name() string
}
