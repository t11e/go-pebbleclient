package pebbleclient

import (
	"errors"
	"fmt"
	"reflect"
)

// TODO: Do we have any methods we can define?
type Service interface{}

type ServiceFactoryFunc func(client Client) (Service, error)

type Registry struct {
	services map[reflect.Type]ServiceFactoryFunc
}

func NewRegistry() *Registry {
	return &Registry{
		services: map[reflect.Type]ServiceFactoryFunc{},
	}
}

func (registry *Registry) Register(intf reflect.Type, fn ServiceFactoryFunc) {
	registry.services[intf] = fn
}

func (registry *Registry) GetFactoryFunc(ptr interface{}) (ServiceFactoryFunc, error) {
	ptrV := reflect.ValueOf(ptr)
	if !ptrV.Elem().CanSet() {
		return nil, errors.New("Argument is not settable")
	}
	t := ptrV.Type()
	for svcType, fn := range registry.services {
		if t.AssignableTo(svcType) {
			return fn, nil
		}
	}
	return nil, fmt.Errorf("No registered service matching type %s", t.String())
}
