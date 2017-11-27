package finder

import "finder-go/errors"

type Service struct {
	name    string
	version string
	list    map[string]string
}

type ServiceFinder struct {

}

type OnServiceUpdateEvent func(string, Service) int

func (f *ServiceFinder) RegisterService(name string, addr string) (error) {
	err := new(errors.FinderError)
	return err
}

func (f *ServiceFinder) UnRegisterService(name string) error {
	return nil
}

func (f *ServiceFinder) UseService(name []string) ([]Service, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ServiceFinder) UseAndSubscribeService(name []string, event OnServiceUpdateEvent) ([]Service, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	return nil
}


