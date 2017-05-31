package pebbleclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ernesto-jimenez/httplogger"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

//go:generate go run vendor/github.com/vektra/mockery/cmd/mockery/mockery.go -name=Client -case=underscore

// Options contains options for the client.
type Options struct {
	// ServiceName of target application.
	ServiceName string

	// APIVersion of target application. Defaults to 1.
	APIVersion int

	// Host is the host name, optionally including the port, to connect to.
	Host string

	// Protocol is the HTTP protocol. Defaults to "http".
	Protocol string

	// HTTPClient is an optional HTTP client instance, which will be used
	// instead of the default.
	HTTPClient *http.Client

	// Session is an optional Checkpoint session key.
	Session string

	// RequestID is an optional request ID that can be passed on to the service.
	RequestID string

	// Logger is an optional interface to permit instrumentation. Ignored if
	// passing in a custom HTTP client.
	Logger httplogger.HTTPLogger

	// Ctx is an optional context.
	Ctx context.Context
}

func (o Options) merge(other *Options) Options {
	if other.ServiceName != "" {
		o.ServiceName = other.ServiceName
	}
	if other.APIVersion != 0 {
		o.APIVersion = other.APIVersion
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
	if other.RequestID != "" {
		o.RequestID = other.RequestID
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
	if newOpts.APIVersion == 0 {
		newOpts.APIVersion = 1
	}
	if newOpts.HTTPClient == nil {
		transport := http.DefaultTransport
		if newOpts.Logger != nil {
			transport = httplogger.NewLoggedTransport(http.DefaultTransport, newOpts.Logger)
		}
		newOpts.HTTPClient = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		}
	}
	return &newOpts
}

// Params is a map of query parameters.
type Params map[string]interface{}

// ToValues is a convenience function to generate url.Values from params.
// Conversion to string is done like so:
//
// - If value is nil, use empty string.
//
// - If value is a string, use that.
//
// - If value implements fmt.Stringer, use that.
//
// - Otherwise, use fmt.Sprintf("%v", v).
//
func (p Params) ToValues() url.Values {
	values := url.Values{}
	for k, v := range p {
		if v == nil {
			values.Set(k, "")
		} else if s, ok := v.(string); ok {
			values.Set(k, s)
		} else if s, ok := v.(fmt.Stringer); ok {
			values.Set(k, s.String())
		} else {
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return values
}

// RequestOptions is a set of options that can be applied to a request.
type RequestOptions struct {
	// Params is an optional map of query parameters.
	Params Params
}

type Client interface {
	// GetOptions returns a copy of the current options.
	GetOptions() Options

	// WithOptions returns a new client that has new default client options. Only
	// non-zero values in the options argument will override the client's options.
	WithOptions(opts Options) Client

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

	// Do performs an HTTP request.
	Do(path string, opts *RequestOptions, method string, body io.Reader,
		result interface{}) error
}

type UID string

var uidRe = regexp.MustCompile(`^([^:]+)\:([^\$]+)\$(.+)$`)

func parseUID(in string) (string, string, int, error) {
	matches := uidRe.FindStringSubmatch(in)
	if matches == nil {
		return "", "", 0, errors.Errorf("invalid uid: %s", in)
	}
	class := matches[1]
	path := matches[2]
	nuid, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", "", 0, errors.Errorf("invalid uid: %s", in)
	}
	return class, path, nuid, nil
}

func (uid UID) Class() (string, error) {
	class, _, _, err := parseUID(string(uid))
	return class, err
}

func (uid UID) Path() (string, error) {
	_, path, _, err := parseUID(string(uid))
	return path, err
}

func (uid UID) NUID() (int, error) {
	_, _, nuid, err := parseUID(string(uid))
	return nuid, err
}

// String implements Stringer.
func (uid UID) String() string {
	return string(uid)
}

// Realm returns the realm part of the path.
func (uid UID) Realm() (string, error) {
	path, err := uid.Path()
	if err != nil {
		return "", err
	}
	i := strings.IndexRune(path, '.')
	if i == -1 {
		return path, nil
	}
	return path[:i], nil
}
