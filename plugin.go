package httpfuzz

import (
	"log"
)

// Plugin must be implemented by a plugin to users to hook the request - response transaction.
type Plugin interface {
	Initialize(environment *Environment) error
	OnSuccess(result *Result) error
	Name() string
}

// Environment contains everything a Plugin needs to configure itself.
type Environment struct {
	Arguments map[string][]string
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

// LoadPlugins loads Plugins from binaries on the filesytem.
func LoadPlugins(logger *log.Logger, paths []string, arguments []string) ([]Plugin, error) {
	plugins := []Plugin{}

	// TODO: load plugin arguments from array
	// TODO: load plugins from paths
	// TODO: pass arguments from map to Plugin.Initialize based on Plugin.Name and the arg map

	return plugins, nil
}
