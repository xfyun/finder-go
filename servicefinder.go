package finder

import (
	"finder-go/common"
	"finder-go/errors"
	"finder-go/utils/zkutil"
)

type ServiceFinder struct {
	zkManager *zkutil.ZkManager
}

func (f *ServiceFinder) RegisterService(name string, addr string) error {
	err := new(errors.FinderError)
	return err
}

func (f *ServiceFinder) UnRegisterService(name string) error {
	return nil
}

func (f *ServiceFinder) UseService(name []string) ([]common.Service, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ServiceFinder) UseAndSubscribeService(name []string, event zkutil.OnServiceUpdateEvent) ([]common.Service, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ServiceFinder) UnSubscribeService(name string) error {
	return nil
}
