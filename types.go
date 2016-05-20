package pebbleclient

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/ernesto-jimenez/httplogger"
)

// Options contains options for the client.
type Options struct {
	// ServiceName of target application.
	ServiceName string

	// ApiVersion of target application. Defaults to 1.
	ApiVersion int

	// Host is the host name, optionally including the port, to connect to.
	Host string

	// Protocol is the HTTP protocol. Defaults to "http".
	Protocol string

	// HTTPClient is an optional HTTP client instance, which will be used
	// instead of the default.
	HTTPClient *http.Client

	// Session is an optional Checkpoint session key.
	Session string

	// RequestId is an optional request ID that can be passed on to the service.
	RequestId string

	// Logger is an optional interface to permit instrumentation. Ignored if
	// passing in a custom HTTP client.
	Logger httplogger.HTTPLogger

	// Ctx is an optional context.
	Ctx context.Context
}

func (o Options) Merge(other *Options) Options {
	if other.ServiceName != "" {
		o.ServiceName = other.ServiceName
	}
	if other.ApiVersion != 0 {
		o.ApiVersion = other.ApiVersion
	}
	if other.Host != "" {
		o.Host = other.Host
	}
	if other.Protocol != "" {
		o.Protocol = other.Protocol
	}
	if other.HTTPClient != nil {
		o.HTTPClient = other.HTTPClient
	}
	if other.Session != "" {
		o.Session = other.Session
	}
	if other.RequestId != "" {
		o.RequestId = other.RequestId
	}
	if other.Logger != nil {
		o.Logger = other.Logger
	}
	if other.Ctx != nil {
		o.Ctx = other.Ctx
	}
	return o
}

// applyDefaults
func (o *Options) applyDefaults() *Options {
	newOpts := *o
	if newOpts.Protocol == "" {
		newOpts.Protocol = "http"
	}
	if newOpts.ApiVersion == 0 {
		newOpts.ApiVersion = 1
	}
	if newOpts.HTTPClient == nil {
		transport := http.DefaultTransport
		if newOpts.Logger != nil {
			transport = httplogger.NewLoggedTransport(http.DefaultTransport, newOpts.Logger)
		}
		newOpts.HTTPClient = &http.Client{
			Transport: transport,
		}
	}
	return &newOpts
}

func (o *Options) validate() error {
	if o.Host == "" {
		return errors.New("Host must be specified in options")
	}
	if o.ServiceName == "" {
		return errors.New("Application name must be specified in options")
	}
	return nil
}

// ClientBuilder generates new client instances.
type ClientBuilder interface {
	NewClient(opts Options) Client
}

// RequestOptions is a set of options that can be applied to a request.
type RequestOptions struct {
	// Params is an optional map of query parameters.
	Params url.Values
}

type Client interface {
	// Options returns a new client that has new default client options. Only
	// non-zero values in the options argument will override the client's options.
	Options(opts Options) Client

	// Get performs a GET request and provides the decoded return value in
	// the result argument, unless nil.
	Get(path string, opts *RequestOptions, result interface{}) error

	// Head performs a HEAD request. Its only use is really to check that
	// the resource does not return an error.
	Head(path string, opts *RequestOptions) error

	// Delete performs a DELETE request and provides the decoded return
	// value in the result argument, unless nil.
	Delete(path string, opts *RequestOptions, result interface{}) error

	// Post performs a POST request. The body can be nil if no body is to be
	// sent. Provides the decoded return value in the result argument, unless
	// nil.
	Post(path string, opts *RequestOptions, body io.Reader, result interface{}) error

	// Put performs a PUT request. The body can be nil if no body is to be
	// sent. Provides the decoded return value in the result argument, unless
	// nil.
	Put(path string, opts *RequestOptions, body io.Reader, result interface{}) error
}
