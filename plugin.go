package httpfuzz

import "log"

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	Initialize(environment *Environment) error
	OnSuccess(result *Result) error
	Name() string
}

// Environment contains everything a Plugin needs to configure itself.
type Environment struct {
	Arguments map[string]string
	Logger    *log.Logger
}

// Result is the request, response and associated metadata to be processed by plugins.
type Result struct {
	Request   *Request
	Response  *Response
	Payload   string
	Location  string
	FieldName string
}
