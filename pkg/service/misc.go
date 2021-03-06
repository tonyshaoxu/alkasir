package service

import (
	"errors"
	"strings"
	"sync"

	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

// A list of services to manage and monitor running services
type Services struct {
	sync.RWMutex                     // the mutex
	items        map[string]*Service // the data
}

// Retreive Service by runtime id
func (s *Services) Service(id string) (*Service, bool) {
	var service *Service
	var found bool
	s.RLock()
	if s.items[id] != nil {
		service = s.items[id].copy()
		found = true
	}
	s.RUnlock()
	return service, found
}

// Return list of all registered services
func (s *Services) AllServices() []*Service {
	var services []*Service
	s.RLock()
	for _, s := range s.items {
		if s != nil {
			c := s.copy()
			services = append(services, c)
		}
	}
	s.RUnlock()
	return services
}

// Add a service to list of managed servers
func (s *Services) add(service *Service) (err error) {
	s.Lock()
	id := service.ID
	if s.items[id] != nil {
		err = errors.New("must have unique id")
	} else {
		s.items[id] = service
	}
	s.Unlock()
	return
}

// Remove a service to list of managed servers
func (s *Services) remove(service *Service) (err error) {
	lg.V(5).Infof("removing service %s", service.ID)
	s.Lock()
	defer s.Unlock()
	id := service.ID
	if s.items[id] == nil {
		return errors.New("service not registered, cannot be removed")
	}
	delete(s.items, id)
	lg.V(19).Infof("removed service %s", service.ID)
	return
}

func (s *Services) AllMethods() []*Method {
	methods := make([]*Method, 0)
	s.RLock()
	for _, s := range s.items {
		if s != nil {
			for _, m := range s.Methods.list {
				if m != nil {
					methods = append(methods, m)
				}
			}
		}
	}
	s.RUnlock()
	return methods
}

func (s *Services) Method(id string) *Method {
	if strings.TrimSpace(id) == "" {
		lg.Infoln("trying to fetch method by illegal key")
		return nil
	}
	var method *Method
	allMethods := s.AllMethods()
	for _, m := range allMethods {
		if m.ID == id {
			method = &Method{}
			*method = *m
		}
	}
	return method
}

// ManagedServices is the central list of running services
var ManagedServices = Services{
	items: make(map[string]*Service, 0),
}

// services id generator instance
var serviceIdGen, _ = shared.NewIDGen("service")

// methods id generator instance
var methodIdGen, _ = shared.NewIDGen("method")

// StopAll stops all services, blocks until everything is shut down.
func StopAll() {
	err := ManagedServices.stopAll()
	if err != nil {
		lg.Error(err)
	}

}

// Return list of all registered services
func (s *Services) stopAll() error {
	s.RLock()
	defer s.RUnlock()
	for _, s := range s.items {
		if s != nil {
			lg.V(10).Infof("stopping service %v", s)
			s.Stop()

		}
	}
	for _, s := range s.items {
		if s != nil {
			lg.V(10).Infof("waiting for service to stop: %v", s)
			s.wait()

		}
	}
	return nil
}
