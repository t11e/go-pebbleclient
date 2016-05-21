package pebbleclient_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/t11e/go-pebbleclient"
)

var realms = pebbleclient.RealmsConfig{
	"acme_inc": &pebbleclient.RealmConfig{
		Host:    "example.com",
		Session: "42smurf99",
	},
}

func TestConnector_Connect_success(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	c.Register((*ServiceAInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))

	var actual ServiceAInterface
	assert.NoError(t, c.Connect(&actual))
}

func TestConnector_Connect_multiple(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	c.Register((*ServiceAInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))
	c.Register((*ServiceBInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))

	var actualA ServiceAInterface
	var actualB ServiceBInterface
	assert.NoError(t, c.Connect(&actualA, &actualB))
}

func TestConnector_Connect_noArgs(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)
	assert.NoError(t, c.Connect())
}

func TestConnector_Connect_missing(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	var actual ServiceAInterface
	assert.Error(t, c.Connect(&actual))
}

func TestConnector_Connect_badArg(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	c.Register((*ServiceAInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))

	var actual int
	assert.Error(t, c.Connect(&actual))
}

func TestConnector_WithRequest_success(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	c.Register((*ServiceAInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))

	req, err := http.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	c2, err := c.WithRequest(req)
	assert.NoError(t, err)

	var actual ServiceAInterface
	assert.NoError(t, c2.Connect(&actual))
	svcA, ok := actual.(*ServiceAImpl)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "example.com", svcA.client.GetOptions().Host)
}

func TestConnector_WithRequest_noConfig(t *testing.T) {
	c, err := pebbleclient.NewConnectorFromConfig(realms)
	assert.NoError(t, err)

	c.Register((*ServiceAInterface)(nil),
		pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
			return &ServiceAImpl{client}, nil
		}))

	req, err := http.NewRequest("GET", "http://smurf.com/", bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	_, err = c.WithRequest(req)
	assert.Error(t, err)

	configErr, ok := err.(*pebbleclient.NoHostConfigError)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, "smurf.com", configErr.Host)
}
