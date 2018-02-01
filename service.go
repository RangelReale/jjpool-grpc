package jjpool_grpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/RangelReale/jjdiscovery"
	"google.golang.org/grpc"
)

type Service struct {
	dclient   *jjdiscovery.Client
	service   string
	opts      serviceOptions
	closechan chan bool
	err       error
	m         sync.RWMutex
}

type serviceDialOptsFunc func() ([]grpc.DialOption, error)

type serviceOptions struct {
	dclient      *jjdiscovery.Client
	jjCliOpt     []jjdiscovery.ClientOption
	jjCliGetOpt  []jjdiscovery.ClientGetOption
	dialOptsFunc serviceDialOptsFunc
	addressList  UniqueStringList
	logfunc      LogFunc
}

func NewService(service string, opts ...ServiceOption) (*Service, error) {
	ret := &Service{
		service:   service,
		closechan: make(chan bool),
	}

	for _, opt := range opts {
		opt(&ret.opts)
	}

	if ret.opts.dclient != nil {
		// use passed dclient
		ret.dclient = ret.opts.dclient
		ret.opts.dclient = nil
	} else {
		// create a new dclient
		var err error
		ret.dclient, err = jjdiscovery.NewClient(ret.opts.jjCliOpt...)
		if err != nil {
			return nil, err
		}
	}

	go ret.run()

	return ret, nil
}

func (s *Service) Err() error {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.err
}

func (s *Service) Stop() {
	s.closechan <- true
}

func (s *Service) run() {
	// get first time list
	cs, err := s.dclient.Get(s.service, s.opts.jjCliGetOpt...)
	if err == nil {
		s.update(cs)
	} else {
		s.log(LEVEL_ERROR, fmt.Sprintf("[service.run] error getting initial service %s info: %v", s.service, err))
	}

	for {
		// watch for changes
		s.log(LEVEL_DEBUG, fmt.Sprintf("[service.run] starting watching service %s", s.service))

		s.m.Lock()

		cwatch, err := s.dclient.Watch(s.service, s.opts.jjCliGetOpt...)
		s.err = err
		if err != nil {
			s.log(LEVEL_ERROR, fmt.Sprintf("[service.run] error watching service %s info: %v", s.service, err))
		}
		s.m.Unlock()

		if err == nil {
			// keep reading from watch
			for {
				select {
				case <-s.closechan:
					return
				case cs := <-cwatch.C:
					if cs == nil {
						return
					}
					s.update(cs)
				}
			}
		} else {
			// error, wait 30 seconds for retry
			select {
			case <-s.closechan:
				return
			case <-time.After(time.Second * 30):
				continue
			}
		}

	}
}

func (s *Service) update(service *jjdiscovery.ClientService) {
	s.m.Lock()
	defer s.m.Unlock()

	s.log(LEVEL_INFO, fmt.Sprintf("[service.update] received service %s update", s.service))
}

func (s *Service) log(level LogLevel, msg string) {
	if s.opts.logfunc != nil {
		s.opts.logfunc(level, msg)
	}
}

//
// options
//

type ServiceOption func(options *serviceOptions)

func ServiceDiscovery(dclient *jjdiscovery.Client) ServiceOption {
	return func(o *serviceOptions) {
		o.dclient = dclient
	}
}

func ServiceDiscoveryOpts(opts ...jjdiscovery.ClientOption) ServiceOption {
	return func(o *serviceOptions) {
		o.jjCliOpt = append(o.jjCliOpt, opts...)
	}
}

func ServiceDiscoveryGetOpts(opts ...jjdiscovery.ClientGetOption) ServiceOption {
	return func(o *serviceOptions) {
		o.jjCliGetOpt = append(o.jjCliGetOpt, opts...)
	}
}

func ServiceDialOpts(dialOptsFunc serviceDialOptsFunc) ServiceOption {
	return func(o *serviceOptions) {
		o.dialOptsFunc = dialOptsFunc
	}
}

func ServiceAddress(address ...string) ServiceOption {
	return func(o *serviceOptions) {
		o.addressList = nil
		o.addressList.Add(address...)
	}
}

func ServiceLogFunc(logFunc LogFunc) ServiceOption {
	return func(o *serviceOptions) {
		o.logfunc = logFunc
	}
}
