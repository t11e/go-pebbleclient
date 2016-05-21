package pebbleclient

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type NoHostConfigError struct {
	Host string
}

func (err *NoHostConfigError) Error() string {
	return fmt.Sprintf("No host configuration for host %q", err.Host)
}

type RealmConfig struct {
	Host    string `json:"host" yaml:"host"`
	Session string `json:"session" yaml:"session"`
}

func (config *RealmConfig) ClientOptions() Options {
	return Options{
		Host:    config.Host,
		Session: config.Session,
	}
}

type RealmsConfig map[string]*RealmConfig

func (c RealmsConfig) FindByHost(host string) *RealmConfig {
	for _, config := range c {
		if config.Host == host {
			return config
		}
	}
	return nil
}

type Connector struct {
	realms   RealmsConfig
	registry *Registry
	client   Client
}

func NewConnectorFromConfig(config RealmsConfig) (*Connector, error) {
	client, err := NewHTTPClient(Options{})
	if err != nil {
		return nil, err
	}
	return &Connector{
		realms:   config,
		registry: NewRegistry(),
		client:   client,
	}, nil
}

// Register registers a new service. The type must be a nil interface pointer; e.g.
// (*MyInterface)(nil). The function is a factory function that takes a client, and
// must return a new service instance.
func (connector *Connector) Register(intf interface{}, fn ServiceFactoryFunc) {
	connector.registry.Register(reflect.TypeOf(intf), fn)
}

// WithRequest returns a new connector which inherits settings from a request. See
// HTTPClient.FromHTTPRequest for more information.
func (connector *Connector) WithRequest(req *http.Request) (*Connector, error) {
	var host string
	if hosts, ok := req.Header["X-Forwarded-Host"]; ok && len(hosts) > 0 {
		host = hosts[len(hosts)-1]
	} else {
		host = strings.SplitN(req.Host, ":", 2)[0]
	}
	config := connector.realms.FindByHost(host)
	if config == nil {
		return nil, &NoHostConfigError{host}
	}
	return &Connector{
		realms:   connector.realms,
		registry: connector.registry,
		client:   connector.client.WithOptions(config.ClientOptions()),
	}, nil
}

// Connect finds one or more services. The input arguments must be pointers to
// variables which have the same interface types as those registered with
// Register().
func (connector *Connector) Connect(services ...interface{}) error {
	for i, ptr := range services {
		fn, err := connector.registry.GetFactoryFunc(ptr)
		if err != nil {
			return errors.Wrapf(err, "Could not get service for argument %d", i)
		}
		service, err := fn(connector.client)
		if err != nil {
			return err
		}
		reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(service))
	}
	return nil
}
