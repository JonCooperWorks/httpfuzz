package httpfuzz

import (
	"log"
	"plugin"
	"sync"
	"time"
)

// Listener must be implemented by a plugin to users to hook the request - response transaction.
// The Listen method will be run in its own goroutine, so plugins cannot block the rest of the program, however panics can take down the entire process.
type Listener interface {
	Listen(results <-chan *Result)
}

// Plugin holds a listener and its input channel for us to send requests to.
type Plugin struct {
	Input chan<- *Result
	Listener
}

// InitializerFunc is a go function that should be exported by a function package.
// It should be named "New".
// Your InitializerFunc should return an instance of your Listener with a reference to httpfuzz's logger for consistent logging.
type InitializerFunc func(*log.Logger) (Listener, error)

// Result is the request, response and associated metadata to be processed by plugins.
type Result struct {
	Request     *Request
	Response    *Response
	Payload     string
	Location    string
	FieldName   string
	TimeElapsed time.Duration
}

// PluginBroker handles sending messages to plugins.
type PluginBroker struct {
	Plugins   []*Plugin
	waitGroup sync.WaitGroup
}

// SendResult sends a *Result to all loaded plugins for further processing.
func (p *PluginBroker) SendResult(result *Result) {
	for _, plugin := range p.Plugins {
		plugin.Input <- result
	}
}

// Close closes all plugin chans that are waiting on results.
// Call close only after all results have been sent.
func (p *PluginBroker) Close() {
	for _, plugin := range p.Plugins {
		close(plugin.Input)
	}
}

// LoadPlugins loads Plugins from binaries on the filesytem.
func LoadPlugins(logger *log.Logger, paths []string) (*PluginBroker, error) {
	plugins := []*Plugin{}

	for _, path := range paths {
		plg, err := plugin.Open(path)
		if err != nil {
			return nil, err
		}

		symbol, err := plg.Lookup("New")
		if err != nil {
			return nil, err
		}

		// Go needs this, InitializerFunc is purely for documentation.
		initializer := symbol.(func(*log.Logger) (Listener, error))
		httpfuzzListener, err := initializer(logger)
		if err != nil {
			return nil, err
		}

		input := make(chan *Result)
		httpfuzzPlugin := &Plugin{
			Input:    input,
			Listener: httpfuzzListener,
		}

		// Listen for results in a goroutine for each plugin
		go httpfuzzPlugin.Listen(input)

		plugins = append(plugins, httpfuzzPlugin)
	}

	pluginManager := &PluginBroker{
		Plugins: plugins,
	}
	pluginManager.waitGroup.Add(len(plugins))
	return pluginManager, nil
}
