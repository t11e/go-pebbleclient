package pebbleclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Params map[string]interface{}

type Body struct {
	// Data may be []byte or an io.Reader.
	Data interface{}

	// Optional content type.
	ContentType string
}

func (body *Body) GetReader() (io.Reader, error) {
	switch t := body.Data.(type) {
	case io.Reader:
		return t, nil
	case []byte:
		return bytes.NewReader(t), nil
	default:
		return nil, fmt.Errorf("Body data must be a reader or []byte, got %T", t)
	}
}

// ClientOptions contains options for the client.
type ClientOptions struct {
	// AppName of target application.
	AppName string

	// ApiVersion of target application. Defaults to 1.
	ApiVersion int

	// Host is the host name, optionally including the port, to connect to.
	Host string

	// Protocol is the HTTP protocol. Defaults to "http".
	Protocol string

	// Session is the Checkpoint session key.
	Session string
}

// Client is a client for the Central API.
type Client struct {
	host       string
	session    string
	protocol   string
	appName    string
	apiVersion int
	httpClient *http.Client
}

// New constructs a new client.
func New(options ClientOptions) (*Client, error) {
	if options.Host == "" {
		return nil, errors.New("Host must be specified in options")
	}
	if options.AppName == "" {
		return nil, errors.New("Application name must be specified in options")
	}

	version := options.ApiVersion
	if version == 0 {
		version = 1
	}
	return &Client{
		host:       options.Host,
		session:    options.Session,
		protocol:   "http",
		apiVersion: version,
		appName:    options.AppName,
		httpClient: &http.Client{},
	}, nil
}

// NewFromRequest constructs a new client that inherits the host name, protocol
// and session from an HTTP request. Any options specified will override inferred
// from the request.
func NewFromRequest(options ClientOptions, req *http.Request) (*Client, error) {
	var opts ClientOptions = options
	if opts.Host == "" {
		opts.Host = req.URL.Host
	}
	if opts.Protocol == "" {
		opts.Protocol = req.URL.Scheme
	}
	if opts.Session == "" {
		if cookie, err := req.Cookie("checkpoint.session"); err != nil {
			if err != http.ErrNoCookie {
				return nil, err
			}
			if s := req.URL.Query().Get("session"); s != "" {
				opts.Session = s
			}
		} else {
			opts.Session = cookie.Value
		}
	}
	return New(opts)
}

func (client *Client) GetOptions() ClientOptions {
	return ClientOptions{
		Host:       client.host,
		Session:    client.session,
		ApiVersion: client.apiVersion,
		AppName:    client.appName,
		Protocol:   client.protocol,
	}
}

// Get performs a GET request. The args may include one or more Params (or *Params).
// The result of the request is stored in the last argument. Returns NotFound error
// if not found.
func (client *Client) Get(path string, args ...interface{}) error {
	params, _, out, err := parseVarArgs(args...)
	if err != nil {
		return err
	}
	if err := performJSONRequest(func() (*http.Response, error) {
		return client.httpClient.Get(client.formatUrl(path, params))
	}, out); err != nil {
		if clientErr, ok := err.(*ClientRequestError); ok && clientErr.Response.StatusCode == 404 {
			return NotFound
		}
		return err
	}
	return nil
}

// Post performs a POST request. The args may include one or more Params (or *Params),
// Body (or *Body). The result, if any, is stored in the result argument.
// Returns NotFound error if not found.
func (client *Client) Post(path string, args ...interface{}) error {
	params, body, out, err := parseVarArgs(args...)
	if body == nil {
		return errors.New("Body must be specified")
	}

	reader, err := body.GetReader()
	if err != nil {
		return err
	}
	if err := performJSONRequest(func() (*http.Response, error) {
		return client.httpClient.Post(client.formatUrl(path, params), body.ContentType, reader)
	}, out); err != nil {
		if clientErr, ok := err.(*ClientRequestError); ok && clientErr.Response.StatusCode == 404 {
			return NotFound
		}
		return err
	}
	return nil
}

func (client *Client) formatUrl(path string, params *Params) string {
	if path[0:1] == "/" {
		path = path[1:]
	}
	result := url.URL{
		Scheme: client.protocol,
		Host:   client.host,
		Path:   fmt.Sprintf("/api/%s/v%d/%s", client.appName, client.apiVersion, path),
	}
	if params != nil {
		values := result.Query()
		for key, value := range *params {
			values.Set(key, fmt.Sprintf("%s", value))
		}
		result.RawQuery = values.Encode()
	}
	if client.session != "" {
		result.Query().Set("session", client.session)
	}
	return result.String()
}

func parseVarArgs(args ...interface{}) (*Params, *Body, interface{}, error) {
	if len(args) == 0 {
		return nil, nil, nil, errors.New("Arguments must include a result (interface{})")
	}
	var params *Params
	var body *Body
	var result interface{}
	for i, arg := range args {
		if i == len(args)-1 {
			result = arg
			continue
		}
		switch t := arg.(type) {
		case Params:
			params = &t
		case *Params:
			params = t
		case Body:
			body = &t
		case *Body:
			body = t
		default:
			return nil, nil, nil, fmt.Errorf("Invalid argument; expected Params, got %T", t)
		}
	}
	return params, body, result, nil
}
