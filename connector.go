package pebbleclient

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type NoRealmConfigError struct {
	Realm string
}

func (err *NoRealmConfigError) Error() string {
	return fmt.Sprintf("No configuration for realm %q", err.Realm)
}

type NoHostConfigError struct {
	Host string
}

func (err *NoHostConfigError) Error() string {
	return fmt.Sprintf("No host configuration for host %q", err.Host)
}

type RealmConfig struct {
	// Host for this realm.
	Host string `json:"host" yaml:"host"`

	// Aliases valid for this realm.
	Aliases []string `json:"aliases" yaml:"aliases"`

	// Session key.
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
		for _, alias := range config.Aliases {
			if normalizeHost(alias) == normalizeHost(host) {
				return config
			}
		}
		if normalizeHost(config.Host) == normalizeHost(host) {
			return config
		}
	}
	return nil
}

func normalizeHost(h string) string {
	if strings.HasSuffix(h, ":80") {
		return strings.TrimSuffix(h, ":80")
	}
	return h
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

// WithRealm returns a new connector which inherits settings from a realm.
func (connector *Connector) WithRealm(name string) (*Connector, error) {
	config, ok := connector.realms[name]
	if !ok {
		return nil, &NoRealmConfigError{name}
	}

	return &Connector{
		realms:   connector.realms,
		registry: connector.registry,
		client:   connector.client.WithOptions(config.ClientOptions()),
	}, nil
}

// WithRequest returns a new connector which inherits settings from a request. See
// HTTPClient.FromHTTPRequest for more information.
func (connector *Connector) WithRequest(req *http.Request) (*Connector, error) {
	host, ok := hostFromRequest(req)
	if !ok {
		return nil, &NoHostConfigError{host}
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
func (connector *Connector) Connect(services ...Service) error {
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
