package pebbleclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestNewHTTPClient_validWithDefaults(t *testing.T) {
	client, err := NewHTTPClient(Options{
		Host:        "localhost",
		ServiceName: "frobnitz",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "", client.GetOptions().Session)
	assert.Equal(t, "localhost", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().ServiceName)
}

func TestNewHTTPClient_validWithOptions(t *testing.T) {
	client, err := NewHTTPClient(Options{
		Host:        "localhost",
		Session:     "uio3ui3ui3",
		ServiceName: "frobnitz",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "uio3ui3ui3", client.GetOptions().Session)
	assert.Equal(t, "localhost", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().ServiceName)
}

type Datum struct {
	Message string `json:"message"`
}

func TestClient_Get_plain(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "GET", req.Method)

		w.Header().Set("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))
	assert.NoError(t, err)
	defer server.Close()

	var result *Datum
	err = client.Get("hello", nil, &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_Get_errorStatusCodes(t *testing.T) {
	status := 400
	for status <= 599 {
		msgBytes := []byte("this failed")

		client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
			w.WriteHeader(status)
			w.Write(msgBytes)
		}))
		assert.NoError(t, err)
		defer server.Close()

		err = client.Get("hello", nil, &Datum{})
		if !assert.Error(t, err) {
			return
		}
		if !assert.IsType(t, &RequestError{}, err) {
			return
		}
		reqErr, ok := err.(*RequestError)
		if !assert.True(t, ok) {
			return
		}
		assert.Equal(t, status, reqErr.Resp.StatusCode)
		assert.Equal(t, msgBytes, reqErr.PartialBody)
		assert.Empty(t, reqErr.Options.Params)

		server.Close()

		status++
	}
}

func TestClient_Get_successStatusCodes(t *testing.T) {
	status := 200
	for status <= 299 {
		datum := &Datum{
			Message: "Say hello to my little friend",
		}

		client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)

			encoder := json.NewEncoder(w)
			encoder.Encode(datum)
		}))
		assert.NoError(t, err)
		defer server.Close()

		var result *Datum
		err = client.Get("hello", nil, &result)
		if !assert.NoError(t, err) {
			return
		}
		if status != http.StatusNoContent && status != http.StatusResetContent {
			assert.Equal(t, datum, result)
		}

		server.Close()

		status++
	}
}

func TestClient_Get_withParams(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "json", req.URL.Query().Get("format"))

		w.Header().Set("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))
	assert.NoError(t, err)
	defer server.Close()

	var result *Datum
	err = client.Get("hello", &RequestOptions{
		Params: Params{
			"format": "json",
		},
	}, &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_Get_withBeginningSlashPath(t *testing.T) {
	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		w.WriteHeader(200)
	}))
	assert.NoError(t, err)
	defer server.Close()

	err = client.Get("/hello", nil, nil)
	assert.NoError(t, err)
}

func TestClient_Get_withPathParams(t *testing.T) {
	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/get/drkropotkin", req.URL.Path)
		w.WriteHeader(200)
	}))
	assert.NoError(t, err)
	defer server.Close()

	err = client.Get("/get/:name", &RequestOptions{
		Params: Params{
			"name":   "drkropotkin",
			"format": "json",
		},
	}, nil)
	assert.NoError(t, err)
}

func TestClient_Get_badContentTypeInResponse(t *testing.T) {
	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("<html></html>"))
	}))
	assert.NoError(t, err)
	defer server.Close()

	var result *Datum
	err = client.Get("hello", nil, &result)
	assert.Error(t, err)
}

func TestClient_Get_withLogging(t *testing.T) {
	logger := &MockLogger{}

	client, server, err := newClientAndServerWithOpts(Options{
		Logger: logger,
	}, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		time.Sleep(50e+6 * time.Nanosecond)
		w.WriteHeader(200)
		w.Write([]byte(`{"answer":42}`))
	}))
	assert.NoError(t, err)
	defer server.Close()

	err = client.Get("hello", &RequestOptions{}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, logger.loggedReq)
	assert.Equal(t, "/api/frobnitz/v1/hello", logger.loggedReq.URL.Path)
	assert.NotNil(t, logger.loggedResp)
	assert.Equal(t, 200, logger.loggedResp.StatusCode)
	assert.Nil(t, logger.loggedErr)
	assert.True(t, logger.loggedDuration >= 50e+6)
}

func TestClient_Post_plain(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	client, server, err := newClientAndServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var mediaType string
		mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		assert.NoError(t, err)
		assert.Equal(t, "application/json", mediaType)

		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "POST", req.Method)

		b, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, []byte(`{"message":"Say hello to my little friend"}`), b)

		w.Header().Set("Content-Type", "application/json")

		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))
	assert.NoError(t, err)
	defer server.Close()

	b, err := json.Marshal(datum)
	assert.NoError(t, err)

	var result *Datum
	err = client.Post("hello", &RequestOptions{}, bytes.NewReader(b), &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_FromHTTPRequest_cookie(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:  "checkpoint.session",
		Value: "uio3ui3ui3",
	})

	client, err := NewHTTPClient(Options{
		ServiceName: "frobnitz",
		Host:        "localhost",
	})
	assert.NoError(t, err)

	client, err = client.FromHTTPRequest(req)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "uio3ui3ui3", client.GetOptions().Session)
	assert.Equal(t, "example.com", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().ServiceName)
}

func TestClient_FromHTTPRequest_sessionParam(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/?session=uio3ui3ui3", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	client, err := NewHTTPClient(Options{
		ServiceName: "frobnitz",
		Host:        "localhost",
	})
	assert.NoError(t, err)

	client, err = client.FromHTTPRequest(req)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "uio3ui3ui3", client.GetOptions().Session)
	assert.Equal(t, "example.com", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().ServiceName)
}
