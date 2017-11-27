package finder

import (
	"finder-go/errors"
)

type Config struct {
	name    string
	file    []byte
	oldFile []byte
}

type ConfigFinder struct {

}

type OnCfgUpdateEvent func(Config) int

func (f *ConfigFinder) UseConfig(name []string) ([]Config, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ConfigFinder) UseAndSubscribeConfig(name []string, event OnCfgUpdateEvent) ([]Config, error) {
	err := new(errors.FinderError)
	return nil, err
}

func (f *ConfigFinder) UnSubscribeConfig(name string) error {
	return nil
}
