package httpfuzz

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	OnSuccess(result *Result) error
	Name() string
}

// Result is the request, response and associated metadata to be processed by plugins.
type Result struct {
	Request   *Request
	Response  *Response
	Payload   string
	Location  string
	FieldName string
}
