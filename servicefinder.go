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

func (f *ServiceFinder) RegisterService(name string) (int, error) {
	err := new(errors.FinderError)
	return Service_Success, err
}

func (f *ServiceFinder) UnRegisterService(name string) error {
	return nil
}

func (f *ServiceFinder) SubscribeService(name []string, event OnServiceUpdateEvent) error {
	err := new(errors.FinderError)
	return err
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	return nil
}


