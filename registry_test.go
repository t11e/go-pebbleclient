package pebbleclient_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t11e/go-pebbleclient"
)

func TestRegistry_success(t *testing.T) {
	svcA := &ServiceAImpl{}
	svcFn := pebbleclient.ServiceFactoryFunc(func(client pebbleclient.Client) (pebbleclient.Service, error) {
		return svcA, nil
	})

	registry := pebbleclient.NewRegistry()
	registry.Register(reflect.TypeOf((*ServiceAInterface)(nil)), svcFn)

	var intf ServiceAInterface
	actualFn, err := registry.GetFactoryFunc(&intf)
	if !assert.NoError(t, err) {
		return
	}
	actualSvc, err := actualFn(nil)
	assert.NoError(t, err)
	assert.Equal(t, svcA, actualSvc)
}

func TestRegistry_multiple(t *testing.T) {
	svcA := &ServiceAImpl{}
	fnA := func(client pebbleclient.Client) (pebbleclient.Service, error) {
		return svcA, nil
	}

	svcB := &ServiceBImpl{}
	fnB := func(client pebbleclient.Client) (pebbleclient.Service, error) {
		return svcB, nil
	}

	registry := pebbleclient.NewRegistry()
	registry.Register(reflect.TypeOf((*ServiceAInterface)(nil)), fnA)
	registry.Register(reflect.TypeOf((*ServiceBInterface)(nil)), fnB)

	var intfA ServiceAInterface
	actualFnA, err := registry.GetFactoryFunc(&intfA)
	if !assert.NoError(t, err) {
		return
	}
	actualSvcA, err := actualFnA(nil)
	assert.NoError(t, err)
	assert.Equal(t, svcA, actualSvcA)

	var intfB ServiceBInterface
	actualFnB, err := registry.GetFactoryFunc(&intfB)
	if !assert.NoError(t, err) {
		return
	}
	actualSvcB, err := actualFnB(nil)
	assert.NoError(t, err)
	assert.Equal(t, svcB, actualSvcB)
}

func TestRegistry_missing(t *testing.T) {
	registry := pebbleclient.NewRegistry()
	var intf ServiceAInterface
	_, err := registry.GetFactoryFunc(&intf)
	assert.Error(t, err)
}

func TestRegistry_badArg(t *testing.T) {
	fn := func(client pebbleclient.Client) (pebbleclient.Service, error) {
		return &ServiceAImpl{}, nil
	}
	registry := pebbleclient.NewRegistry()
	registry.Register(reflect.TypeOf((*ServiceAInterface)(nil)), fn)

	var intf int
	_, err := registry.GetFactoryFunc(&intf)
	assert.Error(t, err)
}
