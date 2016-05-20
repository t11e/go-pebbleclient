package pebbleclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const maxPartialBody = 64 * 1024

type options Options

func (o options) Merge(other *options) options {
	op := Options(*other)
	return options(Options(o).Merge(&op))
}

// HTTPClient is a client for the Central API.
type HTTPClient struct {
	options
	hc *http.Client
}

// New constructs a new client.
func NewHTTPClient(opts Options) (*HTTPClient, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	var newOpts options = *(*options)(opts.ApplyDefaults())
	return &HTTPClient{
		options: newOpts,
		hc:      newOpts.HTTPClient,
	}, nil
}

// NewFromHTTPRequest constructs a new client that inherits the host name, protocol,
// session and request ID from an HTTP request. Any options specified will override inferred
// from the request.
func (client *HTTPClient) FromHTTPRequest(req *http.Request) (*HTTPClient, error) {
	opts := Options(client.options)

	var host string
	if hosts, ok := req.Header["X-Forwarded-Host"]; ok && len(hosts) > 0 {
		host = hosts[len(hosts)-1]
	} else {
		host = strings.SplitN(req.Host, ":", 2)[0]
	}
	opts.Host = host

	opts.Protocol = req.URL.Scheme

	if session := req.URL.Query().Get("session"); session != "" {
		opts.Session = session
	} else if cookie, err := req.Cookie("checkpoint.session"); err == nil {
		opts.Session = cookie.Value
	}

	if id := req.Header.Get("Request-Id"); id != "" {
		opts.RequestId = id
	}

	opts.HTTPClient = client.hc

	return NewHTTPClient(opts)
}

// GetOptions returns a copy of this client's options.
func (client *HTTPClient) GetOptions() Options {
	return Options(client.options)
}

func (client *HTTPClient) Options(opts Options) Client {
	newOpts := client.options.Merge((*options)(&opts))
	return &HTTPClient{
		options: newOpts,
		hc:      client.hc,
	}
}

func (client *HTTPClient) Get(path string, opts *RequestOptions, result interface{}) error {
	return client.do(opts, http.MethodGet, path, nil, result)
}

func (client *HTTPClient) Head(path string, opts *RequestOptions) error {
	return client.do(opts, http.MethodHead, path, nil, nil)
}

func (client *HTTPClient) Post(path string, opts *RequestOptions, body io.Reader, result interface{}) error {
	return client.do(opts, http.MethodPost, path, body, result)
}

func (client *HTTPClient) Put(path string, opts *RequestOptions, body io.Reader, result interface{}) error {
	return client.do(opts, http.MethodPut, path, body, result)
}

func (client *HTTPClient) Delete(path string, opts *RequestOptions, result interface{}) error {
	return client.do(opts, http.MethodDelete, path, nil, result)
}

func (client *HTTPClient) do(
	opts *RequestOptions,
	method string,
	path string,
	bodyIn io.Reader,
	result interface{}) error {
	if opts == nil {
		opts = &RequestOptions{}
	}

	req, err := http.NewRequest(method, client.formatEndpointUrl(path, opts.Params), bodyIn)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf8")
	if client.options.RequestId != "" {
		req.Header.Set("Request-Id", client.options.RequestId)
	}

	ctx := client.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := ctxhttp.Do(ctx, client.hc, req)
	if err != nil {
		return err
	}

	bodyOut := resp.Body
	defer func() {
		if bodyOut != nil {
			_ = bodyOut.Close()
		}
	}()

	if isNonSuccessStatus(resp.StatusCode) {
		return client.buildError(&RequestError{}, opts, resp)
	}

	if doesStatusCodeYieldBody(resp.StatusCode) && result != nil {
		return decodeResponseAsJSON(resp, bodyOut, result)
	}

	return nil
}

func (client *HTTPClient) buildError(
	err *RequestError,
	opts *RequestOptions,
	resp *http.Response) error {
	var buf bytes.Buffer
	b := make([]byte, 1024)
	for buf.Len() < maxPartialBody {
		count, err := resp.Body.Read(b[:])
		if count == 0 {
			break
		}
		if err != nil && err != io.EOF {
			break
		}
		_, wErr := buf.Write(b[0:count])
		if err != nil || wErr != nil {
			break
		}
	}
	err.PartialBody = buf.Bytes()
	err.client = client
	err.Resp = resp
	err.Options = opts
	return err
}

func (client *HTTPClient) formatEndpointUrl(path string, params Params) string {
	if path[0:1] == "/" {
		path = path[1:]
	}
	result := url.URL{
		Scheme: client.Protocol,
		Host:   client.Host,
		Path: fmt.Sprintf("/api/%s/v%d/%s",
			client.ServiceName, client.ApiVersion, escapedPath(path)),
	}

	query := result.Query()
	if params != nil {
		for key, value := range params {
			query.Set(key, fmt.Sprintf("%s", value))
		}
	}
	if client.Session != "" {
		query.Set("session", client.Session)
	}
	result.RawQuery = query.Encode()

	return result.String()
}
