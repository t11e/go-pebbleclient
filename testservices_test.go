package pebbleclient_test

import "github.com/t11e/go-pebbleclient"

type ServiceAInterface interface {
	Ping() error
}

type ServiceAImpl struct {
	client pebbleclient.Client
}

func (m *ServiceAImpl) Ping() error {
	return m.client.Get("/ping", nil, nil)
}

type ServiceBInterface interface {
	Ping() error
}

type ServiceBImpl struct {
	client pebbleclient.Client
}

func (m *ServiceBImpl) Ping() error {
	return m.client.Get("/ping", nil, nil)
}
