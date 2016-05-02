package pebbleclient

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_validWithDefaults(t *testing.T) {
	client, err := New(ClientOptions{
		Host:    "localhost",
		AppName: "frobnitz",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "", client.GetOptions().Session)
	assert.Equal(t, "localhost", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().AppName)
}

func TestNew_validWithOptions(t *testing.T) {
	client, err := New(ClientOptions{
		Host:    "localhost",
		Session: "uio3ui3ui3",
		AppName: "frobnitz",
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "uio3ui3ui3", client.GetOptions().Session)
	assert.Equal(t, "localhost", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().AppName)
}

func TestNew_invalidHost(t *testing.T) {
	_, err := New(ClientOptions{Host: ""})
	assert.Error(t, err)
}

func TestNewFromRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "checkpoint.session",
		Value: "uio3ui3ui3",
	})

	client, err := NewFromRequest(ClientOptions{
		AppName: "frobnitz",
	}, req)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "uio3ui3ui3", client.GetOptions().Session)
	assert.Equal(t, "example.com", client.GetOptions().Host)
	assert.Equal(t, "http", client.GetOptions().Protocol)
	assert.Equal(t, "frobnitz", client.GetOptions().AppName)
}

type Datum struct {
	Message string `json:"message"`
}

func TestClient_Get(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))

	client, err := New(ClientOptions{
		Host:    hostFromUrl(server.URL),
		AppName: "frobnitz",
	})
	assert.NoError(t, err)

	var result *Datum
	err = client.Get("hello", &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_Get_params(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "json", req.URL.Query().Get("format"))
		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))

	client, err := New(ClientOptions{
		Host:    hostFromUrl(server.URL),
		AppName: "frobnitz",
	})
	assert.NoError(t, err)

	var result *Datum
	err = client.Get("hello", Params{"format": "json"}, &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_Get_notFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		w.WriteHeader(404)
	}))

	client, err := New(ClientOptions{
		Host:    hostFromUrl(server.URL),
		AppName: "frobnitz",
	})
	assert.NoError(t, err)

	var result *Datum
	err = client.Get("hello", &result)
	assert.Equal(t, NotFound, err)
}

func TestClient_Post_bytes(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	b, err := json.Marshal(datum)
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))

	client, err := New(ClientOptions{
		Host:    hostFromUrl(server.URL),
		AppName: "frobnitz",
	})
	assert.NoError(t, err)

	var result *Datum
	err = client.Post("hello", Body{Data: b, ContentType: "application/json"}, &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}

func TestClient_Post_reader(t *testing.T) {
	datum := &Datum{
		Message: "Say hello to my little friend",
	}

	b, err := json.Marshal(datum)
	assert.NoError(t, err)

	reader := bytes.NewReader(b)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/frobnitz/v1/hello", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		encoder := json.NewEncoder(w)
		encoder.Encode(datum)
	}))

	client, err := New(ClientOptions{
		Host:    hostFromUrl(server.URL),
		AppName: "frobnitz",
	})
	assert.NoError(t, err)

	var result *Datum
	err = client.Post("hello", Body{Data: reader, ContentType: "application/json"}, &result)
	assert.NoError(t, err)
	assert.Equal(t, datum, result)
}
