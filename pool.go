package jjpool_grpc

import (
	"sync"

	"github.com/RangelReale/jjdiscovery"
	"google.golang.org/grpc"
)

type Pool struct {
	dclient  *jjdiscovery.Client
	opts     poolOptions
	services map[string]*Service
	m        sync.RWMutex
}

type dialOptsFunc func(service string) ([]grpc.DialOption, error)

type poolOptions struct {
	jjCliOpt     []jjdiscovery.ClientOption
	dialOptsFunc dialOptsFunc
	logfunc      LogFunc
}

func NewPool(opts ...PoolOption) (*Pool, error) {
	ret := &Pool{}

	for _, opt := range opts {
		opt(&ret.opts)
	}

	var err error
	ret.dclient, err = jjdiscovery.NewClient(ret.opts.jjCliOpt...)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *Pool) AddService(service string, opts ...ServiceOption) error {
	p.m.Lock()
	defer p.m.Unlock()

	var nopts []ServiceOption
	if p.opts.logfunc != nil {
		// logfunc must be first, to not possibly override the user
		nopts = append(nopts, ServiceLogFunc(p.opts.logfunc))
		nopts = append(nopts, opts...)
	} else {
		nopts = opts
	}
	nopts = append(nopts, ServiceDiscovery(p.dclient))

	// creates the service
	s, err := NewService(service, nopts...)
	if err != nil {
		return err
	}
	p.services[service] = s

	return nil
}

//
// options
//

type PoolOption func(options *poolOptions)

func PoolDiscoveryOpts(opts ...jjdiscovery.ClientOption) PoolOption {
	return func(o *poolOptions) {
		o.jjCliOpt = append(o.jjCliOpt, opts...)
	}
}

func PoolDialOpts(dialOptsFunc dialOptsFunc) PoolOption {
	return func(o *poolOptions) {
		o.dialOptsFunc = dialOptsFunc
	}
}

func PooolLogFunc(logFunc LogFunc) PoolOption {
	return func(o *poolOptions) {
		o.logfunc = logFunc
	}
}
