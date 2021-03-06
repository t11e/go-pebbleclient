package pebbleclient

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jpillora/backoff"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

const maxPartialBody = 64 * 1024

type options Options

func (o options) merge(other *options) options {
	op := Options(*other)
	return options(Options(o).merge(&op))
}

// HTTPClient is a client for the Central API.
type HTTPClient struct {
	options
	hc *http.Client
}

// NewHTTPClient constructs a new client.
func NewHTTPClient(opts Options) (*HTTPClient, error) {
	o := *(*options)(opts.applyDefaults())
	return &HTTPClient{
		options: o,
		hc:      o.HTTPClient,
	}, nil
}

// FromHTTPRequest constructs a new client that inherits the host name, protocol,
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
		opts.RequestID = id
	}

	opts.HTTPClient = client.hc

	return NewHTTPClient(opts)
}

func (client *HTTPClient) GetOptions() Options {
	return Options(client.options)
}

func (client *HTTPClient) WithOptions(opts Options) Client {
	newOpts := client.options.merge((*options)(&opts))
	return &HTTPClient{
		options: newOpts,
		hc:      client.hc,
	}
}

func (client *HTTPClient) Get(path string, opts *RequestOptions, result interface{}) error {
	return client.Do(path, opts, "GET", nil, result)
}

func (client *HTTPClient) Head(path string, opts *RequestOptions) error {
	return client.Do(path, opts, "HEAD", nil, nil)
}

func (client *HTTPClient) Post(path string, opts *RequestOptions, body io.Reader, result interface{}) error {
	return client.Do(path, opts, "POST", body, result)
}

func (client *HTTPClient) Put(path string, opts *RequestOptions, body io.Reader, result interface{}) error {
	return client.Do(path, opts, "PUT", body, result)
}

func (client *HTTPClient) Delete(path string, opts *RequestOptions, result interface{}) error {
	return client.Do(path, opts, "DELETE", nil, result)
}

func (client *HTTPClient) Do(
	path string,
	opts *RequestOptions,
	method string,
	body io.Reader,
	result interface{}) error {
	if client.Host == "" {
		return errors.New("Host name not configured")
	}

	if client.ServiceName == "" {
		return errors.New("Application name not configured")
	}

	if opts == nil {
		opts = &RequestOptions{}
	}

	url, err := client.formatEndpointURL(path, opts.Params)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if client.options.RequestID != "" {
		req.Header.Set("Request-Id", client.options.RequestID)
	}
	if client.Session != "" {
		req.AddCookie(&http.Cookie{
			Name:  "checkpoint.session",
			Value: client.Session,
		})
	}

	ctx := client.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	boff := &backoff.Backoff{
		Jitter: true,
	}

	for {
		resp, err := ctxhttp.Do(ctx, client.hc, req)
		if err != nil {
			return err
		}

		respBody := resp.Body
		defer func() {
			if respBody != nil {
				// Drain remaining body to work around bug in Go < 1.7
				_, _ = io.Copy(ioutil.Discard, respBody)

				_ = respBody.Close()
			}
		}()

		if isNonSuccessStatus(resp.StatusCode) {
			if isRetriableStatus(resp.StatusCode) {
				select {
				case <-ctx.Done():
					// Let error handling below handle this
				case <-time.After(boff.Duration()):
					continue
				}
			}
			return client.buildError(&RequestError{}, opts, req, resp)
		}

		if doesStatusCodeYieldBody(resp.StatusCode) && result != nil {
			return decodeResponseAsJSON(resp, respBody, result)
		}
		return nil
	}
}

func (client *HTTPClient) buildError(
	error *RequestError,
	opts *RequestOptions,
	req *http.Request,
	resp *http.Response) error {
	b, _ := ioutil.ReadAll(&io.LimitedReader{
		R: resp.Body,
		N: maxPartialBody,
	})
	error.PartialBody = b
	error.client = client
	error.Req = req
	error.Resp = resp
	error.Options = opts
	return error
}

func (client *HTTPClient) formatEndpointURL(path string, params Params) (string, error) {
	values := params.ToValues()

	var err error
	path, err = formatPath(path, values)
	if err != nil {
		return "", err
	}
	if path[0:1] == "/" {
		path = path[1:]
	}
	result := url.URL{
		Scheme: client.Protocol,
		Host:   client.Host,
		Path:   fmt.Sprintf("/api/%s/v%d/%s", client.ServiceName, client.APIVersion, path),
	}

	query := result.Query()
	if values != nil {
		for k, vs := range values {
			for _, v := range vs {
				query.Add(k, v)
			}
		}
	}
	result.RawQuery = query.Encode()

	return result.String(), nil
}
