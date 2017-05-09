package pebbleclient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	pebbleclient "github.com/t11e/go-pebbleclient"
)

func hostFromUrl(anUrl string) string {
	u, err := url.Parse(anUrl)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL %q", anUrl))
	}
	return u.Host
}

func newClientAndServer(serverHandler http.HandlerFunc) (pebbleclient.Client, *httptest.Server, error) {
	return newClientAndServerWithOpts(pebbleclient.Options{}, serverHandler)
}

func newClientAndServerWithOpts(
	opts pebbleclient.Options,
	serverHandler http.HandlerFunc) (pebbleclient.Client, *httptest.Server, error) {
	server := httptest.NewServer(serverHandler)

	var newOpts pebbleclient.Options = opts
	if newOpts.Host == "" {
		newOpts.Host = hostFromUrl(server.URL)
	}
	if newOpts.ServiceName == "" {
		newOpts.ServiceName = "frobnitz"
	}

	client, err := pebbleclient.NewHTTPClient(newOpts)
	return client, server, err
}

type MockLogger struct {
	loggedReq      *http.Request
	loggedResp     *http.Response
	loggedErr      error
	loggedDuration time.Duration
}

func (l *MockLogger) LogRequest(req *http.Request) {
	l.loggedReq = req
}

func (l *MockLogger) LogResponse(req *http.Request, resp *http.Response, err error, duration time.Duration) {
	l.loggedResp = resp
	l.loggedErr = err
	l.loggedDuration = duration
}
